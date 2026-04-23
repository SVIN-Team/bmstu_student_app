//go:build unit

package module

import (
	"context"
	"testing"
	"time"

	"stud_hub/internal/config"
	autherrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	"stud_hub/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthUseCase_SignUp_Success(t *testing.T) {
	ctx := context.Background()
	cfg := config.AuthConfig{
		AccessSecretKey:  "test-access-secret",
		RefreshSecretKey: "test-refresh-secret",
		AccessLifeTime:   15 * time.Minute,
		RefreshLifeTime:  7 * 24 * time.Hour,
	}

	user := models.User{
		Email:        "test@example.com",
		PasswordHash: "plain_password",
	}

	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)

	expectedUserID := uuid.New()
	mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("models.User")).Return(expectedUserID, nil)
	mockTokenRepo.On("SaveRefreshToken", ctx, mock.AnythingOfType("models.RefreshToken")).Return(nil)

	authUC := usecase.NewAuthUseCase(mockTokenRepo, mockUserRepo, cfg)

	accessToken, refreshToken, userID, err := authUC.SignUp(ctx, user)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, expectedUserID, userID)

	mockUserRepo.AssertExpectations(t)
    mockTokenRepo.AssertExpectations(t)
}

func TestAuthUseCase_SignUp_DuplicateEmail(t *testing.T) {
	ctx := context.Background()
	cfg := config.AuthConfig{
		AccessSecretKey:  "test-access-secret",
		RefreshSecretKey: "test-refresh-secret",
		AccessLifeTime:   15 * time.Minute,
		RefreshLifeTime:  7 * 24 * time.Hour,
	}

	user := models.User{
		Email:        "duplicate@example.com",
		PasswordHash: "plain_password",
	}

	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)

	mockUserRepo.On("CreateUser", ctx, mock.AnythingOfType("models.User")).Return(uuid.Nil, autherrors.ErrUserDuplicate)

	authUC := usecase.NewAuthUseCase(mockTokenRepo, mockUserRepo, cfg)

	accessToken, refreshToken, userID, err := authUC.SignUp(ctx, user)

	assert.Error(t, err)
	assert.Equal(t, autherrors.ErrUserDuplicate, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Equal(t, uuid.Nil, userID)
}

func TestAuthUseCase_SignIn_Success(t *testing.T) {
	ctx := context.Background()
	cfg := config.AuthConfig{
		AccessSecretKey:  "test-access-secret",
		RefreshSecretKey: "test-refresh-secret",
		AccessLifeTime:   15 * time.Minute,
		RefreshLifeTime:  7 * 24 * time.Hour,
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("plain_password"), bcrypt.DefaultCost)

	existingUser := models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	loginUser := models.User{
		Email:        "test@example.com",
		PasswordHash: "plain_password",
	}

	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)

	mockUserRepo.On("GetUserByEmail", ctx, loginUser.Email).Return(existingUser, nil)
	mockTokenRepo.On("DeleteUserRefreshTokens", ctx, existingUser.ID).Return(nil)
	mockTokenRepo.On("SaveRefreshToken", ctx, mock.AnythingOfType("models.RefreshToken")).Return(nil)

	authUC := usecase.NewAuthUseCase(mockTokenRepo, mockUserRepo, cfg)

	accessToken, refreshToken, userID, err := authUC.SignIn(ctx, loginUser)

	assert.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, refreshToken)
	assert.Equal(t, existingUser.ID, userID)
}

func TestAuthUseCase_SignIn_InvalidPassword(t *testing.T) {
	ctx := context.Background()
	cfg := config.AuthConfig{
		AccessSecretKey:  "test-access-secret",
		RefreshSecretKey: "test-refresh-secret",
		AccessLifeTime:   15 * time.Minute,
		RefreshLifeTime:  7 * 24 * time.Hour,
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct_password"), bcrypt.DefaultCost)

	existingUser := models.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
	}

	loginUser := models.User{
		Email:        "test@example.com",
		PasswordHash: "wrong_password",
	}

	mockUserRepo := new(MockUserRepository)
	mockTokenRepo := new(MockTokenRepository)

	mockUserRepo.On("GetUserByEmail", ctx, loginUser.Email).Return(existingUser, nil)

	authUC := usecase.NewAuthUseCase(mockTokenRepo, mockUserRepo, cfg)

	accessToken, refreshToken, userID, err := authUC.SignIn(ctx, loginUser)

	assert.Error(t, err)
	assert.Equal(t, autherrors.ErrInvalidCredentials, err)
	assert.Empty(t, accessToken)
	assert.Empty(t, refreshToken)
	assert.Equal(t, uuid.Nil, userID)
}