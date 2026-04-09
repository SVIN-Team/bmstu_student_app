package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"stud_hub/internal/config"
	errors2 "stud_hub/internal/errors"
	"stud_hub/internal/handler/http/dto"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AuthUseCase interface {
	SignUp(ctx context.Context, user models.User) (accessToken string, refreshToken string, userId uuid.UUID, err error)
	SignIn(ctx context.Context, user models.User) (accessToken string, refreshToken string, userId uuid.UUID, err error)
	Refresh(ctx context.Context, refreshTokenString string) (accessToken string, newRefreshToken string, err error)
	SignOut(ctx context.Context, refreshTokenString string) error
	DeleteUserTokens(ctx context.Context, userID uuid.UUID) error
}

type GroupUseCase interface {
	GetAll(ctx context.Context) ([]models.Group, error)
	GetByID(ctx context.Context, id uuid.UUID) (models.Group, error)
	GetByName(ctx context.Context, name string) (models.Group, error)
	Create(ctx context.Context, group models.Group) (uuid.UUID, error)
	Update(ctx context.Context, group models.Group) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type AuthHandler struct {
	authUseCase  AuthUseCase
	groupUseCase GroupUseCase
	authConf     *config.AuthConfig
}

func (h *AuthHandler) isCookieSecure() bool {
	return !h.authConf.InsecureCookies
}

func NewAuthHandler(authUseCase AuthUseCase, groupUseCase GroupUseCase, cfg *config.AuthConfig) *AuthHandler {
	if cfg == nil {
		panic(errors.New("auth config is nil"))
	}
	return &AuthHandler{
		authUseCase:  authUseCase,
		groupUseCase: groupUseCase,
		authConf:     cfg,
	}
}

// SignUp godoc
// @Summary Register a new user
// @Description Creates a new user in an existing group and returns created user payload.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.SignUpRequest true "Signup request"
// @Success 200 {object} dto.SuccessResponse{data=dto.AuthSignUpResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/register [post]
func (h *AuthHandler) SignUp(ctx *gin.Context) {
	var json dto.SignUpRequest
	err := ctx.ShouldBindBodyWithJSON(&json)
	if err != nil {
		ValidationError(ctx, fmt.Sprintf("could not parse body: %v", err))
		logger.Infof(ctx, "Failed to parse signup request body: %v", err)
		return
	}

	group, err := h.groupUseCase.GetByName(ctx, json.GroupName)
	if errors.Is(err, errors2.ErrGroupNotFound) {
		NotFoundError(ctx, fmt.Sprintf("group %s not found", json.GroupName))
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("could not get group: %v", err))
		return
	}

	userModel := json.ToModel(group.ID)

	access, refresh, userId, err := h.authUseCase.SignUp(ctx, userModel)
	if err != nil {
		// TODO: classify errors
		InternalError(ctx, fmt.Sprintf("could not sign up: %v", err))
		return
	}

	secure := h.isCookieSecure()
	ctx.SetCookie("access_token", access, int(h.authConf.AccessLifeTime.Seconds()),
		"/", "", secure, true)
	ctx.SetCookie("refresh_token", refresh, int(h.authConf.RefreshLifeTime.Seconds()),
		"/api/v1/auth/refresh", "", secure, true)

	SuccessResponse(ctx, http.StatusOK, dto.AuthSignUpResponse{
		ID:        userId.String(),
		Email:     json.Email,
		FirstName: json.FirstName,
		LastName:  json.LastName,
	})
}

// SignIn godoc
// @Summary Sign in
// @Description Authenticates user credentials and sets auth cookies.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Signin request"
// @Success 200 {object} dto.SuccessResponse{data=dto.AuthSignInResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/login [post]
func (h *AuthHandler) SignIn(ctx *gin.Context) {
	var json dto.LoginRequest
	err := ctx.ShouldBindBodyWithJSON(&json)
	if err != nil {
		ValidationError(ctx, fmt.Sprintf("could not parse body: %v", err))
		logger.Infof(ctx, "Failed to parse signin request body: %v", err)
		return
	}

	access, refresh, userId, err := h.authUseCase.SignIn(ctx, json.ToModel())
	if errors.Is(err, errors2.ErrUserNotFound) {
		NotFoundError(ctx, fmt.Sprintf("user %s not found", json.ToModel().Email))
		logger.Infof(ctx, "Failed to sign in: %v", err)
		return
	} else if errors.Is(err, errors2.ErrInvalidCredentials) {
		UnauthorizedError(ctx, fmt.Sprintf("invalid credentials: %v", err))
		logger.Infof(ctx, "Failed to sign in: %v", err)
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("could not sign in: %v", err))
		logger.Errorf(ctx, "Failed to sign in: %v", err)
		return
	}

	secure := h.isCookieSecure()
	ctx.SetCookie("access_token", access, int(h.authConf.AccessLifeTime.Seconds()),
		"/", "", secure, true)
	ctx.SetCookie("refresh_token", refresh, int(h.authConf.RefreshLifeTime.Seconds()),
		"/api/v1/auth/refresh", "", secure, true)
	SuccessResponse(ctx, http.StatusOK, dto.AuthSignInResponse{
		ID: userId.String(),
	})
}

// Refresh godoc
// @Summary Refresh access token
// @Description Refreshes access and refresh tokens using refresh cookie.
// @Tags Auth
// @Produce json
// @Success 200 {object} dto.SuccessResponse{data=dto.AuthRefreshResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		UnauthorizedError(ctx, "No refresh token provided")
		return
	}

	access, newRefresh, err := h.authUseCase.Refresh(ctx, refreshToken)
	if errors.Is(err, errors2.ErrInvalidRefreshToken) {
		ErrorResponse(ctx, http.StatusUnauthorized, "INVALID_TOKEN", "Invalid or expired refresh token")
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to refresh token: %v", err))
		logger.Errorf(ctx, "Failed to refresh token: %v", err)
		return
	}

	secure := h.isCookieSecure()
	ctx.SetCookie("access_token", access, int(h.authConf.AccessLifeTime.Seconds()),
		"/", "", secure, true)
	ctx.SetCookie("refresh_token", newRefresh, int(h.authConf.RefreshLifeTime.Seconds()),
		"/api/v1/auth/refresh", "", secure, true)

	SuccessResponse(ctx, http.StatusOK, dto.AuthRefreshResponse{
		AccessToken: access,
		ExpiresIn:   int(h.authConf.AccessLifeTime.Seconds()),
	})
}

// SignOut godoc
// @Summary Sign out from current device
// @Description Revokes current refresh token and clears auth cookies.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout [post]
func (h *AuthHandler) SignOut(ctx *gin.Context) {
	refreshToken, err := ctx.Cookie("refresh_token")
	if err != nil {
		UnauthorizedError(ctx, "No refresh token provided")
		return
	}

	err = h.authUseCase.SignOut(ctx, refreshToken)
	if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to sign out: %v", err))
		logger.Errorf(ctx, "Failed to sign out: %v", err)
		return
	}

	secure := h.isCookieSecure()
	ctx.SetCookie("access_token", "", -1, "/", "", secure, true)
	ctx.SetCookie("refresh_token", "", -1, "/api/v1/auth/refresh", "", secure, true)

	SuccessResponse(ctx, http.StatusOK, dto.MessageResponse{
		Message: "Successfully signed out",
	})
}

// SignOutAll godoc
// @Summary Sign out from all devices
// @Description Revokes all refresh tokens for the authenticated user and clears auth cookies.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.MessageResponse}
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /auth/logout-all [post]
func (h *AuthHandler) SignOutAll(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	userIDValue, exists := ctx.Get("user_id")
	if !exists {
		UnauthorizedError(ctx, "User not authenticated")
		return
	}

	userID, ok := userIDValue.(uuid.UUID)
	if !ok {
		InternalError(ctx, "Invalid user ID in context")
		return
	}

	err := h.authUseCase.DeleteUserTokens(ctx, userID)
	if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to sign out from all devices: %v", err))
		logger.Errorf(ctx, "Failed to sign out from all devices: %v", err)
		return
	}

	secure := h.isCookieSecure()
	ctx.SetCookie("access_token", "", -1, "/", "", secure, true)
	ctx.SetCookie("refresh_token", "", -1, "/api/v1/auth/refresh", "", secure, true)

	SuccessResponse(ctx, http.StatusOK, dto.MessageResponse{
		Message: "Successfully signed out from all devices",
	})
}
