package usecase

import (
	"context"
	"errors"

	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
)

type GroupRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Group, error)
	GetByName(ctx context.Context, name string) (models.Group, error)
	GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error)
	GetAll(ctx context.Context) ([]models.Group, error)
	Create(ctx context.Context, group models.Group) (uuid.UUID, error)
	Update(ctx context.Context, group models.Group) error
	Delete(ctx context.Context, id uuid.UUID) error
	HasUsers(ctx context.Context, id uuid.UUID) (bool, error)
}

type GroupUseCase struct {
	groupRepo GroupRepository
}

func NewGroupUseCase(groupRepo GroupRepository) *GroupUseCase {
	return &GroupUseCase{groupRepo: groupRepo}
}

func (g *GroupUseCase) GetAll(ctx context.Context) ([]models.Group, error) {
	groups, err := g.groupRepo.GetAll(ctx)
	if err != nil {
		logger.Errorf(ctx, "failed to get groups: %v", err)
		return nil, apperrors.ErrInternalServer
	}
	return groups, nil
}

func (g *GroupUseCase) GetByID(ctx context.Context, id uuid.UUID) (models.Group, error) {
	group, err := g.groupRepo.GetByID(ctx, id)
	if err != nil {
		return models.Group{}, apperrors.ErrGroupNotFound
	}
	return group, nil
}

func (g *GroupUseCase) Create(ctx context.Context, group models.Group) (uuid.UUID, error) {
	existing, err := g.groupRepo.GetByName(ctx, group.Name)
	if err != nil {
		if !errors.Is(err, apperrors.ErrGroupNotFound) {
			logger.Errorf(ctx, "failed to get group by name: %v", err)
			return uuid.Nil, apperrors.ErrInternalServer
		}
	} else if existing.ID != uuid.Nil {
		return uuid.Nil, apperrors.ErrGroupAlreadyExists
	}

	group.ID = uuid.New()

	id, err := g.groupRepo.Create(ctx, group)
	if err != nil {
		logger.Errorf(ctx, "failed to create group: %v", err)
		return uuid.Nil, apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "group created: %s (%s)", id, group.Name)
	return id, nil
}

func (g *GroupUseCase) Update(ctx context.Context, group models.Group) error {
	if _, err := g.groupRepo.GetByID(ctx, group.ID); err != nil {
		return apperrors.ErrGroupNotFound
	}

	if err := g.groupRepo.Update(ctx, group); err != nil {
		logger.Errorf(ctx, "failed to update group: %v", err)
		if errors.Is(err, apperrors.ErrUniqueViolationFault) {
			return apperrors.ErrUniqueViolationFault
		}
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "group updated: %s", group.ID)
	return nil
}

func (g *GroupUseCase) Delete(ctx context.Context, id uuid.UUID) error {
	if _, err := g.groupRepo.GetByID(ctx, id); err != nil {
		return apperrors.ErrGroupNotFound
	}

	hasUsers, err := g.groupRepo.HasUsers(ctx, id)
	if err != nil {
		logger.Errorf(ctx, "failed to check users: %v", err)
		return apperrors.ErrInternalServer
	}
	if hasUsers {
		return apperrors.ErrGroupHasUsers
	}

	if err := g.groupRepo.Delete(ctx, id); err != nil {
		logger.Errorf(ctx, "failed to delete group: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "group deleted: %s", id)
	return nil
}

func (g *GroupUseCase) GetByName(ctx context.Context, name string) (models.Group, error) {
	return g.groupRepo.GetByName(ctx, name)
}
