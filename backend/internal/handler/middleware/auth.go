package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"stud_hub/internal/config"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Context keys for storing user information
type contextKey string

const (
	UserIDContextKey   contextKey = "user_id"
	UserRoleContextKey contextKey = "user_role"
)

// UserReader interface for fetching user data
type UserReader interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
}

// AuthMiddleware struct holds the configuration for authentication
type AuthMiddleware struct {
	authConfig config.AuthConfig
	userReader UserReader
}

// NewAuthMiddleware creates a new instance of AuthMiddleware
func NewAuthMiddleware(authConfig config.AuthConfig, userReader UserReader) *AuthMiddleware {
	return &AuthMiddleware{
		authConfig: authConfig,
		userReader: userReader,
	}
}

// AuthRequired is a middleware that validates JWT access tokens
func (m *AuthMiddleware) AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get the access token from the cookie
		tokenString, err := c.Cookie("access_token")
		if err != nil {
			// If not in cookie, try to get it from Authorization header
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "UNAUTHORIZED",
						"message": "No access token provided",
					},
				})
				c.Abort()
				return
			}

			// Check if the header has the Bearer prefix
			if !strings.HasPrefix(authHeader, "Bearer ") {
				c.JSON(http.StatusUnauthorized, gin.H{
					"success": false,
					"error": gin.H{
						"code":    "UNAUTHORIZED",
						"message": "Invalid authorization header format",
					},
				})
				c.Abort()
				return
			}

			// Extract the token
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}

		// Parse and validate the token
		claims, err := m.parseAccessToken(c, tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid or expired access token",
				},
			})
			c.Abort()
			return
		}

		// Extract user ID from claims
		userID, err := uuid.Parse(claims.Subject)
		if err != nil {
			logger.Warnf(c, "Failed to parse user ID from token claims: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "INVALID_TOKEN",
					"message": "Invalid user ID in token",
				},
			})
			c.Abort()
			return
		}

		// Add user ID to context
		c.Set(string(UserIDContextKey), userID)

		// Continue with the next handler
		c.Next()
	}
}

// AuthRequiredWithRole is a middleware that validates JWT and loads user role into context
func (m *AuthMiddleware) AuthRequiredWithRole() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run the AuthRequired middleware to validate token
		m.AuthRequired()(c)

		// If the previous middleware aborted, don't continue
		if c.IsAborted() {
			return
		}

		// Get user ID from context
		userID, err := GetUserIDFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User not authenticated",
				},
			})
			c.Abort()
			return
		}

		// Fetch user from database to get their role
		user, err := m.userReader.GetUserByID(c, userID)
		if err != nil {
			logger.Errorf(c, "Failed to fetch user %s: %v", userID, err)
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User not found",
				},
			})
			c.Abort()
			return
		}

		// Check if user is blocked
		if user.IsBlocked {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "User account is blocked",
				},
			})
			c.Abort()
			return
		}

		// Add user role to context
		c.Set(string(UserRoleContextKey), user.Role)

		// Continue with the next handler
		c.Next()
	}
}

// AdminRequired is a middleware that ensures the user has admin role
func (m *AuthMiddleware) AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run the AuthRequiredWithRole middleware
		m.AuthRequiredWithRole()(c)

		// If the previous middleware aborted, don't continue
		if c.IsAborted() {
			return
		}

		// Get user role from context
		role, err := GetUserRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User role not found",
				},
			})
			c.Abort()
			return
		}

		// Check if user is admin
		if role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Admin access required",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// HeadmanRequired is a middleware that ensures the user has headman role
func (m *AuthMiddleware) HeadmanRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run the AuthRequiredWithRole middleware
		m.AuthRequiredWithRole()(c)

		// If the previous middleware aborted, don't continue
		if c.IsAborted() {
			return
		}

		// Get user role from context
		role, err := GetUserRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User role not found",
				},
			})
			c.Abort()
			return
		}

		// Check if user is headman
		if role != models.RoleHeadman {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Headman access required",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireRole is a middleware that checks if the user has one of the required roles
func (m *AuthMiddleware) RequireRole(requiredRoles ...models.RoleType) gin.HandlerFunc {
	return func(c *gin.Context) {
		// First run the AuthRequiredWithRole middleware
		m.AuthRequiredWithRole()(c)

		// If the previous middleware aborted, don't continue
		if c.IsAborted() {
			return
		}

		// Get user role from context
		role, err := GetUserRoleFromContext(c)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "UNAUTHORIZED",
					"message": "User role not found",
				},
			})
			c.Abort()
			return
		}

		// Check if user has one of the required roles
		hasRole := false
		for _, requiredRole := range requiredRoles {
			if role == requiredRole {
				hasRole = true
				break
			}
		}

		if !hasRole {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"error": gin.H{
					"code":    "FORBIDDEN",
					"message": "Insufficient permissions",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// parseAccessToken parses and validates a JWT access token
func (m *AuthMiddleware) parseAccessToken(ctx context.Context, tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate the signing method
		if token.Method != jwt.SigningMethodHS256 {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.authConfig.AccessSecretKey), nil
	})

	if err != nil {
		logger.Warnf(ctx, "Failed to parse access token: %v", err)
		return nil, err
	}

	if !token.Valid {
		logger.Warnf(ctx, "Invalid access token")
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		logger.Warnf(ctx, "Failed to parse token claims")
		return nil, errors.New("failed to parse claims")
	}

	return claims, nil
}

// GetUserIDFromContext extracts the user ID from the gin context
func GetUserIDFromContext(c *gin.Context) (uuid.UUID, error) {
	userIDValue, exists := c.Get(string(UserIDContextKey))
	if !exists {
		return uuid.Nil, errors.New("user ID not found in context")
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		return uuid.Nil, errors.New("invalid user ID in context")
	}

	return userID, nil
}

// GetUserRoleFromContext extracts the user role from the gin context
func GetUserRoleFromContext(c *gin.Context) (models.RoleType, error) {
	roleValue, exists := c.Get(string(UserRoleContextKey))
	if !exists {
		return "", errors.New("user role not found in context")
	}

	role, ok := roleValue.(models.RoleType)
	if !ok {
		return "", errors.New("invalid user role in context")
	}

	return role, nil
}
