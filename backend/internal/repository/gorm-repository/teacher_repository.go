package gormrepository

import (
	"context"

	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	gormmodels "stud_hub/internal/repository/gorm-models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type TeacherRepository struct {
	db *gorm.DB
}

func NewTeacherRepository(db *gorm.DB) *TeacherRepository {
	return &TeacherRepository{db: db}
}

func (r *TeacherRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Teacher, error) {
	var t gormmodels.Teacher
	if err := r.db.WithContext(ctx).First(&t, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			logger.Warnf(ctx, "gorm: teacher with id %s not found: %v", id, err)
			return models.Teacher{}, apperrors.ErrTeacherNotFound
		}
		logger.Errorf(ctx, "gorm: failed to get teacher %s: %v", id, err)
		return models.Teacher{}, apperrors.ErrInternalServer
	}
	return models.Teacher{
		ID:         t.ID,
		FirstName:  t.FirstName,
		LastName:   t.LastName,
		Patronymic: t.Patronymic,
	}, nil
}

func (r *TeacherRepository) GetOrCreateByFullName(ctx context.Context, lastName, firstName, patronymic string) (uuid.UUID, error) {
	var t gormmodels.Teacher
	err := r.db.WithContext(ctx).Where("first_name = ?", firstName).Where("last_name = ?", lastName).Where("patronymic = ?", patronymic).First(&t).Error
	if err == nil {
		return t.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		logger.Errorf(ctx, "gorm: failed to find teacher %s: %v", lastName, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}

	t = gormmodels.Teacher{LastName: lastName, FirstName: firstName, Patronymic: patronymic}
	if err := r.db.WithContext(ctx).Create(&t).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create teacher %s: %v", lastName, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}
	return t.ID, nil
}