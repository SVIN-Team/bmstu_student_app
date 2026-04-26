package gormrepository

import (
	"context"
	"errors"
	"sync"

	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	gormmodels "stud_hub/internal/repository/gorm-models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
	mu sync.Mutex
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db, mu: sync.Mutex{}}
}

func (r *GroupRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Group, error) {
	var g gormmodels.Group
	if err := r.db.WithContext(ctx).First(&g, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Group{}, apperrors.ErrGroupNotFound
		}
		logger.Errorf(ctx, "gorm: failed to get group %s: %v", id, err)
		return models.Group{}, apperrors.ErrInternalServer
	}
	return models.Group{ID: g.ID, Name: g.Name}, nil
}

func (r *GroupRepository) GetByName(ctx context.Context, name string) (models.Group, error) {
	var g gormmodels.Group
	if err := r.db.WithContext(ctx).First(&g, "name = ?", name).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return models.Group{}, apperrors.ErrGroupNotFound
		}
		logger.Errorf(ctx, "gorm: failed to get group by name %s: %v", name, err)
		return models.Group{}, apperrors.ErrInternalServer
	}
	return models.Group{ID: g.ID, Name: g.Name}, nil
}

func (r *GroupRepository) GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	var g gormmodels.Group
	err := r.db.WithContext(ctx).First(&g, "name = ?", name).Error
	if err == nil {
		return g.ID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.Errorf(ctx, "gorm: failed to find group by name %s: %v", name, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}

	g = gormmodels.Group{Name: name}
	if err := r.db.WithContext(ctx).Create(&g).Error; err != nil {
		if isUniqueViolationFault(err) {
			if err := r.db.WithContext(ctx).First(&g, "name = ?", name).Error; err == nil {
				return g.ID, nil
			}
		}
		logger.Errorf(ctx, "gorm: failed to create group %s: %v", name, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}

	return g.ID, nil
}

func (r *GroupRepository) GetAll(ctx context.Context) ([]models.Group, error) {
	var groups []gormmodels.Group
	if err := r.db.WithContext(ctx).Find(&groups).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to list groups: %v", err)
		return nil, apperrors.ErrInternalServer
	}

	result := make([]models.Group, 0, len(groups))
	for _, g := range groups {
		result = append(result, models.Group{ID: g.ID, Name: g.Name})
	}
	return result, nil
}

func (r *GroupRepository) Create(ctx context.Context, group models.Group) (uuid.UUID, error) {
	g := gormmodels.Group{
		ID:   group.ID,
		Name: group.Name,
	}
	if err := r.db.WithContext(ctx).Create(&g).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create group %s: %v", group.Name, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}
	return g.ID, nil
}

func (r *GroupRepository) Update(ctx context.Context, group models.Group) error {
	if err := r.db.WithContext(ctx).Model(&gormmodels.Group{ID: group.ID}).Update("name", group.Name).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to update group %s: %v", group.ID, err)
		if isUniqueViolationFault(err) {
			return apperrors.ErrUniqueViolationFault
		}
		return apperrors.ErrInternalServer
	}
	return nil
}

func (r *GroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&gormmodels.Group{}, "id = ?", id).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to delete group %s: %v", id, err)
		return apperrors.ErrInternalServer
	}
	return nil
}

func (r *GroupRepository) HasUsers(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&gormmodels.User{}).Where("group_id = ?", id).Count(&count).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to count users for group %s: %v", id, err)
		return false, apperrors.ErrInternalServer
	}
	return count > 0, nil
}
