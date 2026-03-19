package gormrepository

import (
	"context"

	autherrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	gormmodels "stud_hub/internal/repository/gorm-models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository implements usecase.UserRepository using GORM.
type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	gUser := toGormUser(user)

	if err := r.db.WithContext(ctx).Create(&gUser).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create user %s: %v", user.Email, err)
		return uuid.UUID{}, autherrors.ErrInternalServer
	}
	return gUser.ID, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, uid uuid.UUID) (models.User, error) {
	var gUser gormmodels.User
	if err := r.db.WithContext(ctx).First(&gUser, "id = ?", uid).Error; err != nil {
		logger.Warnf(ctx, "gorm: failed to get user by id %s: %v", uid, err)
		return models.User{}, err
	}
	return fromGormUser(gUser), nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user models.User) (models.User, error) {
	gUser := toGormUser(user)

	builder := PartialUpdateBuilder()
	builder.
		UpdateField("email", gUser.Email).
		UpdateField("first_name", gUser.FirstName).
		UpdateField("last_name", gUser.LastName).
		UpdateField("patronymic", gUser.Patronymic).
		UpdateValue("role", gUser.Role).
		SetPtrUUID("group_id", gUser.GroupID).
		UpdateValue("is_blocked", user.IsBlocked)
		

	if gUser.PasswordHash != nil && *gUser.PasswordHash != "" {
		builder.UpdateValue("password_hash", gUser.PasswordHash)
	}

	updates := builder.Build()

	if len(updates) == 0 {
		return r.GetUserByID(ctx, user.ID)
	}

	if err := r.db.WithContext(ctx).
		Model(&gormmodels.User{ID: user.ID}).
		Updates(updates).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to update user %s: %v", user.ID, err)
		return models.User{}, autherrors.ErrInternalServer
	}

	return r.GetUserByID(ctx, user.ID)
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var gUser gormmodels.User
	if err := r.db.WithContext(ctx).First(&gUser, "email = ?", email).Error; err != nil {
		logger.Warnf(ctx, "gorm: failed to get user by email %s: %v", email, err)
		return models.User{}, err
	}
	return fromGormUser(gUser), nil
}

func toGormUser(u models.User) gormmodels.User {
	var password *string
	if u.PasswordHash != "" {
		password = &u.PasswordHash
	}

	return gormmodels.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: password,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Patronymic:   u.Patronymic,
		Role:         gormmodels.UserRole(u.Role),
		IsBlocked:    u.IsBlocked,
		GroupID:      nullableUUID(u.GroupID),
		CreatedAt:    u.CreatedAt,
	}
}

func fromGormUser(u gormmodels.User) models.User {
	var groupID uuid.UUID
	if u.GroupID != nil {
		groupID = *u.GroupID
	}
	var password string
	if u.PasswordHash != nil {
		password = *u.PasswordHash
	}

	return models.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: password,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		Role:         models.RoleType(u.Role),
		Patronymic:   u.Patronymic,
		GroupID:      groupID,
		IsBlocked:    u.IsBlocked,
		CreatedAt:    u.CreatedAt,
	}
}

func nullableUUID(id uuid.UUID) *uuid.UUID {
	if id == uuid.Nil {
		return nil
	}
	return &id
}
