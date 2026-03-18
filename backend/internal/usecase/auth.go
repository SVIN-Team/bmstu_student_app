package usecase

import (
	"context"
	"errors"
	"stud_hub/internal/config"
	autherrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	"stud_hub/util/logger"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type TokenRepository interface {
	SaveRefreshToken(ctx context.Context, token models.RefreshToken) error
	GetRefreshToken(ctx context.Context, jti uuid.UUID) (models.RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, jti uuid.UUID) error
	DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error
}

type UserRepository interface {
	CreateUser(ctx context.Context, user models.User) (uuid.UUID, error)
	GetUserByID(ctx context.Context, uid uuid.UUID) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) (models.User, error)
	GetUserByEmail(ctx context.Context, email string) (models.User, error)
}

type AuthUseCase struct {
	tokenRepo TokenRepository
	userRepo  UserRepository
	authCfg   config.AuthConfig
}

func NewAuthUseCase(tokenRepo TokenRepository, userRepo UserRepository, cfg config.AuthConfig) *AuthUseCase {
	return &AuthUseCase{
		tokenRepo: tokenRepo,
		userRepo:  userRepo,
		authCfg:   cfg,
	}
}

func (a *AuthUseCase) SignUp(ctx context.Context, user models.User) (accessToken string, refreshToken string, err error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		logger.Errorf(ctx, "failed to hash password: %v", err)
		return "", "", autherrors.ErrInternalServer
	}
	user.PasswordHash = string(hashedPassword)

	uid, err := a.userRepo.CreateUser(ctx, user)
	if err != nil {
		logger.Errorf(ctx, "failed to create user: %v", err)
		return "", "", autherrors.ErrInternalServer
	}

	accessToken, refreshToken, err = a.createTokenPair(ctx, uid)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (a *AuthUseCase) SignIn(ctx context.Context, user models.User) (accessToken string, refreshToken string, err error) {
	userData, err := a.userRepo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		logger.Warnf(ctx, "failed to get user by email %s: %v", user.Email, err)
		return "", "", autherrors.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(userData.PasswordHash), []byte(user.PasswordHash))
	if err != nil {
		logger.Warnf(ctx, "invalid password for user %s", user.Email)
		return "", "", autherrors.ErrInvalidCredentials
	}

	err = a.tokenRepo.DeleteUserRefreshTokens(ctx, userData.ID)
	if err != nil {
		logger.Errorf(ctx, "failed to delete old user refresh tokens for user %s: %v", userData.ID, err)
	}

	accessToken, refreshToken, err = a.createTokenPair(ctx, userData.ID)
	if err != nil {
		return "", "", err
	}
	return accessToken, refreshToken, nil
}

func (a *AuthUseCase) Refresh(ctx context.Context, refreshTokenString string) (accessToken string, newRefreshToken string, err error) {
	claims, err := a.parseRefreshToken(refreshTokenString)
	if err != nil {
		return "", "", err
	}

	jti, err := uuid.Parse(claims.ID)
	if err != nil {
		logger.Warnf(ctx, "failed to parse JTI from refresh token: %v", err)
		return "", "", autherrors.ErrInvalidRefreshToken
	}

	_, err = a.tokenRepo.GetRefreshToken(ctx, jti)
	if err != nil {
		logger.Warnf(ctx, "refresh token with jti %s not found in db: %v", jti, err)
		return "", "", autherrors.ErrInvalidRefreshToken
	}

	err = a.tokenRepo.DeleteRefreshToken(ctx, jti)
	if err != nil {
		logger.Errorf(ctx, "failed to delete old refresh token %s: %v", jti, err)
		return "", "", autherrors.ErrInternalServer
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		logger.Errorf(ctx, "failed to parse user ID from refresh token claims: %v", err)
		return "", "", autherrors.ErrInternalServer
	}

	accessToken, newRefreshToken, err = a.createTokenPair(ctx, userID)
	if err != nil {
		return "", "", err
	}

	return accessToken, newRefreshToken, nil
}

func (a *AuthUseCase) SignOut(ctx context.Context, refreshTokenString string) error {
	claims, err := a.parseRefreshToken(refreshTokenString)
	if err != nil {
		return err
	}

	jti, err := uuid.Parse(claims.ID)
	if err != nil {
		logger.Warnf(ctx, "failed to parse JTI from refresh token on logout: %v", err)
		return autherrors.ErrInvalidRefreshToken
	}

	err = a.tokenRepo.DeleteRefreshToken(ctx, jti)
	if err != nil {
		logger.Errorf(ctx, "failed to delete refresh token on logout: %v", err)
		return autherrors.ErrInternalServer
	}

	return nil
}

func (a *AuthUseCase) DeleteUserTokens(ctx context.Context, userID uuid.UUID) error {
	err := a.tokenRepo.DeleteUserRefreshTokens(ctx, userID)
	if err != nil {
		logger.Errorf(ctx, "failed to delete user refresh tokens for user %s: %v", userID, err)
		return autherrors.ErrInternalServer
	}
	return nil
}

func (a *AuthUseCase) createTokenPair(ctx context.Context, userID uuid.UUID) (string, string, error) {
	access, err := a.generateAccessToken(userID)
	if err != nil {
		logger.Errorf(ctx, "failed to create access token for user %s: %v", userID, err)
		return "", "", autherrors.ErrInternalServer
	}

	refresh, tokenModel, err := a.generateRefreshToken(userID)
	if err != nil {
		logger.Errorf(ctx, "failed to create refresh token for user %s: %v", userID, err)
		return "", "", autherrors.ErrInternalServer
	}

	err = a.tokenRepo.SaveRefreshToken(ctx, tokenModel)
	if err != nil {
		logger.Errorf(ctx, "failed to save refresh token for user %s: %v", userID, err)
		return "", "", autherrors.ErrInternalServer
	}
	return access, refresh, nil
}

func (a *AuthUseCase) generateAccessToken(userID uuid.UUID) (string, error) {
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Issuer:    "stud_hub",
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.authCfg.AccessLifeTime)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.authCfg.AccessSecretKey))
}

func (a *AuthUseCase) generateRefreshToken(userID uuid.UUID) (string, models.RefreshToken, error) {
	jti := uuid.New()
	issuedAt := time.Now()
	expiresAt := issuedAt.Add(a.authCfg.RefreshLifeTime)

	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		Issuer:    "stud_hub",
		IssuedAt:  jwt.NewNumericDate(issuedAt),
		ExpiresAt: jwt.NewNumericDate(expiresAt),
		ID:        jti.String(), // JTI claim
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	ss, err := token.SignedString([]byte(a.authCfg.RefreshSecretKey))
	if err != nil {
		return "", models.RefreshToken{}, err
	}

	return ss, models.RefreshToken{
		ID:        jti,
		UserID:    userID,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
	}, nil
}

func (a *AuthUseCase) parseRefreshToken(ctx context.Context, tokenString string) (*jwt.RegisteredClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(a.authCfg.RefreshSecretKey), nil
	})

	if err != nil || !token.Valid {
		logger.Warnf(ctx, "invalid refresh token provided: %v", err)
		return nil, autherrors.ErrInvalidRefreshToken
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return nil, autherrors.ErrInvalidRefreshToken
	}

	return claims, nil
}
