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

type SubjectRepository struct {
	db *gorm.DB
}

func NewSubjectRepository(db *gorm.DB) *SubjectRepository {
	return &SubjectRepository{db: db}
}

func (r *SubjectRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Subject, error) {
	var subj gormmodels.Subject
	if err := r.db.WithContext(ctx).First(&subj, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Subject{}, apperrors.ErrSubjectNotFound
		}
		logger.Errorf(ctx, "gorm: failed to get subject %s: %v", id, err)
		return models.Subject{}, apperrors.ErrInternalServer
	}
	return models.Subject{ID: subj.ID, Name: subj.Name}, nil
}

func (r *SubjectRepository) GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error) {
	var subj gormmodels.Subject
	err := r.db.WithContext(ctx).First(&subj, "name = ?", name).Error
	if err == nil {
		return subj.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		logger.Errorf(ctx, "gorm: failed to find subject by name %s: %v", name, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}

	subj = gormmodels.Subject{Name: name}
	if err := r.db.WithContext(ctx).Create(&subj).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create subject %s: %v", name, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}
	return subj.ID, nil
}
