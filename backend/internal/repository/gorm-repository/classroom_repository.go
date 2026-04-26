package gormrepository

import (
	"context"
	"sync"

	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	gormmodels "stud_hub/internal/repository/gorm-models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ClassroomRepository struct {
	db *gorm.DB
	mu sync.Mutex
}

func NewClassroomRepository(db *gorm.DB) *ClassroomRepository {
	return &ClassroomRepository{db: db, mu: sync.Mutex{}}
}

func (r *ClassroomRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Classroom, error) {
	var room gormmodels.Room
	if err := r.db.WithContext(ctx).First(&room, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Classroom{}, apperrors.ErrClassroomNotFound
		}
		logger.Errorf(ctx, "gorm: failed to get classroom %s: %v", id, err)
		return models.Classroom{}, apperrors.ErrInternalServer
	}
	return models.Classroom{ID: room.ID, Name: derefString(room.Name)}, nil
}

func (r *ClassroomRepository) GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error) {
	var room gormmodels.Room
	r.mu.Lock()
	defer r.mu.Unlock()

	err := r.db.WithContext(ctx).First(&room, "name = ?", name).Error
	if err == nil {
		return room.ID, nil
	}
	if err != gorm.ErrRecordNotFound {
		logger.Errorf(ctx, "gorm: failed to find classroom %s: %v", name, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}

	room = gormmodels.Room{Name: &name}
	if err := r.db.WithContext(ctx).Create(&room).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create classroom %s: %v", name, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}
	return room.ID, nil
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
