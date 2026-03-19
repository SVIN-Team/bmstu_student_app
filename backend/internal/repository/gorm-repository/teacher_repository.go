package gormrepository

import (
	"context"
	"strings"

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

	first, last, patronymic := splitFullName(t.FullName)
	return models.Teacher{
		ID:         t.ID,
		FirstName:  first,
		LastName:   last,
		Patronymic: patronymic,
	}, nil
}

func (r *TeacherRepository) GetOrCreateByFullName(ctx context.Context, lastName, firstName, patronymic string) (uuid.UUID, error) {
	fullName := buildFullName(lastName, firstName, patronymic)

	var t gormmodels.Teacher
	err := r.db.WithContext(ctx).First(&t, "full_name = ?", fullName).Error
	if err == nil {
		return t.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		logger.Errorf(ctx, "gorm: failed to find teacher %s: %v", fullName, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}

	t = gormmodels.Teacher{FullName: fullName}
	if err := r.db.WithContext(ctx).Create(&t).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create teacher %s: %v", fullName, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}
	return t.ID, nil
}

func buildFullName(last, first, patronymic string) string {
	parts := []string{last, first, patronymic}
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return strings.Join(out, " ")
}

func splitFullName(full string) (first, last, patronymic string) {
	parts := strings.Fields(full)
	switch len(parts) {
	case 0:
		return "", "", ""
	case 1:
		return parts[0], parts[0], ""
	case 2:
		return parts[1], parts[0], ""
	default:
		return parts[1], parts[0], strings.Join(parts[2:], " ")
	}
}
