package gormrepository

import (
    "context"
    "errors"

    autherrors "stud_hub/internal/errors"
    "stud_hub/internal/models"
    gormmodels "stud_hub/internal/repository/gorm-models"
    "stud_hub/util/logger"

    "github.com/google/uuid"
    "github.com/jackc/pgconn"
    "gorm.io/gorm"
)

// UserRepository implements usecase.UserRepository using GORM.
type UserRepository struct {
    db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
    return &UserRepository{db: db}
}

func isUniqueViolationFault(err error) bool {
    if err == nil {
        return false
    }

    var pgErr *pgconn.PgError
    if errors.As(err, &pgErr) {
        return pgErr.Code == "23505" // PostgreSQL error code for unique constraint violation.
    }

    return false
}

func (r *UserRepository) CreateUser(ctx context.Context, user models.User) (uuid.UUID, error) {
    gUser := gormmodels.ToGormUser(user)

    if err := r.db.WithContext(ctx).Create(&gUser).Error; err != nil {
        logger.Errorf(ctx, "gorm: failed to create user %s: %v", user.Email, err)
        if isUniqueViolationFault(err) {
            return uuid.UUID{}, autherrors.ErrUserDuplicate
        }
        return uuid.UUID{}, autherrors.ErrInternalServer
    }
    return gUser.ID, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, uid uuid.UUID) (models.User, error) {
    var gUser gormmodels.User
    if err := r.db.WithContext(ctx).First(&gUser, "id = ?", uid).Error; err != nil {
        logger.Warnf(ctx, "gorm: failed to get user by id %s: %v", uid, err)
        return models.User{}, autherrors.ErrUserNotFound
    }
    return gormmodels.FromGormUser(gUser), nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, user models.User) (models.User, error) {
    gUser := gormmodels.ToGormUser(user)

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
        return models.User{}, autherrors.ErrUserNotFound
    }
    return gormmodels.FromGormUser(gUser), nil
}

func (r *UserRepository) GetAllUsers(ctx context.Context, page, perPage int) ([]models.User, int, error) {
    var gUsers []gormmodels.User
    var total int64

    dbQuery := r.db.WithContext(ctx).Model(&gormmodels.User{})

    if err := dbQuery.Count(&total).Error; err != nil {
        logger.Errorf(ctx, "gorm: failed to count users: %v", err)
        return nil, 0, autherrors.ErrInternalServer
    }

    offset := (page - 1) * perPage
    if err := dbQuery.Offset(offset).Limit(perPage).Find(&gUsers).Error; err != nil {
        logger.Errorf(ctx, "gorm: failed to get users: %v", err)
        return nil, 0, autherrors.ErrInternalServer
    }

    var users []models.User
    for _, gu := range gUsers {
        users = append(users, gormmodels.FromGormUser(gu))
    }

    return users, int(total), nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uuid.UUID) error {
    if err := r.db.WithContext(ctx).Delete(&gormmodels.User{}, "id = ?", id).Error; err != nil {
        logger.Errorf(ctx, "gorm: failed to delete user %s: %v", id, err)
        return autherrors.ErrInternalServer
    }
    return nil
}

func (r *UserRepository) BlockUser(ctx context.Context, id uuid.UUID, block bool) error {
    if err := r.db.WithContext(ctx).Model(&gormmodels.User{ID: id}).Update("is_blocked", block).Error; err != nil {
        logger.Errorf(ctx, "gorm: failed to block/unblock user %s: %v", id, err)
        return autherrors.ErrInternalServer
    }
    return nil
}

func (r *UserRepository) ChangeUserRole(ctx context.Context, id uuid.UUID, role models.RoleType) error {
    if err := r.db.WithContext(ctx).Model(&gormmodels.User{ID: id}).Update("role", role).Error; err != nil {
        logger.Errorf(ctx, "gorm: failed to change role for user %s: %v", id, err)
        return autherrors.ErrInternalServer
    }
    return nil
}
