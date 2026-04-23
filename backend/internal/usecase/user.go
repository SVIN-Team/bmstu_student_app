package usecase

import (
	"context"
	"errors"
	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
)

type UserUseCase struct {
	userRepo       UserRepository
	queueSlotsRepo QueueSlotsRepository
	queueRepo      QueueRepository
}

func NewUserUseCase(
	userRepo UserRepository,
	queueSlotsRepo QueueSlotsRepository,
	queueRepo QueueRepository,
) *UserUseCase {
	return &UserUseCase{
		userRepo:       userRepo,
		queueSlotsRepo: queueSlotsRepo,
		queueRepo:      queueRepo,
	}
}

func (u *UserUseCase) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	user, err := u.userRepo.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return models.User{}, apperrors.ErrUserNotFound
		}
		logger.Errorf(ctx, "failed to get user %s: %v", id, err)
		return models.User{}, apperrors.ErrInternalServer
	}
	return user, nil
}

func (u *UserUseCase) UpdateUser(ctx context.Context, user models.User) (models.User, error) {
	// Check if user exists
	existing, err := u.userRepo.GetUserByID(ctx, user.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return models.User{}, apperrors.ErrUserNotFound
		}
		logger.Errorf(ctx, "failed to get user %s: %v", user.ID, err)
		return models.User{}, apperrors.ErrInternalServer
	}

	// Prevent changing critical fields
	user.Email = existing.Email
	user.PasswordHash = existing.PasswordHash
	user.Role = existing.Role
	user.IsBlocked = existing.IsBlocked
	user.CreatedAt = existing.CreatedAt

	updatedUser, err := u.userRepo.UpdateUser(ctx, user)
	if err != nil {
		logger.Errorf(ctx, "failed to update user %s: %v", user.ID, err)
		return models.User{}, apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "user updated: %s", user.ID)
	return updatedUser, nil
}

func (u *UserUseCase) TransferHeadmanRole(ctx context.Context, fromUserID, toUserID uuid.UUID) error {
	// Get the current headman
	fromUser, err := u.userRepo.GetUserByID(ctx, fromUserID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return apperrors.ErrUserNotFound
		}
		logger.Errorf(ctx, "failed to get user %s: %v", fromUserID, err)
		return apperrors.ErrInternalServer
	}

	// Verify the user is headman
	if fromUser.Role != models.RoleHeadman {
		return apperrors.ErrForbidden
	}

	// Get the target user
	toUser, err := u.userRepo.GetUserByID(ctx, toUserID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return apperrors.ErrUserNotFound
		}
		logger.Errorf(ctx, "failed to get user %s: %v", toUserID, err)
		return apperrors.ErrInternalServer
	}

	// Verify both users are in the same group
	if fromUser.GroupID != toUser.GroupID {
		logger.Warnf(ctx, "cannot transfer headman role: users in different groups")
		return apperrors.ErrForbidden
	}

	// Update roles
	fromUser.Role = models.RoleStudent
	toUser.Role = models.RoleHeadman

	if _, err := u.userRepo.UpdateUser(ctx, fromUser); err != nil {
		logger.Errorf(ctx, "failed to update from-user %s: %v", fromUserID, err)
		return apperrors.ErrInternalServer
	}

	if _, err := u.userRepo.UpdateUser(ctx, toUser); err != nil {
		// Rollback the first update
		fromUser.Role = models.RoleHeadman
		if _, rollbackErr := u.userRepo.UpdateUser(ctx, fromUser); rollbackErr != nil {
			logger.Errorf(ctx, "failed to rollback headman role transfer: %v", rollbackErr)
		}
		logger.Errorf(ctx, "failed to update to-user %s: %v", toUserID, err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "headman role transferred from %s to %s", fromUserID, toUserID)
	return nil
}

func (u *UserUseCase) GetUserSlots(ctx context.Context, userID uuid.UUID, statusFilter, queueStatusFilter string) ([]models.QueueSlot, error) {
	// Get user to get their group
	user, err := u.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		if errors.Is(err, apperrors.ErrUserNotFound) {
			return nil, apperrors.ErrUserNotFound
		}
		logger.Errorf(ctx, "failed to get user %s: %v", userID, err)
		return nil, apperrors.ErrInternalServer
	}

	// Get all queues for the user's group
	queues, err := u.queueRepo.GetByGroupID(ctx, user.GroupID)
	if err != nil {
		logger.Errorf(ctx, "failed to get queues for group %s: %v", user.GroupID, err)
		return nil, apperrors.ErrInternalServer
	}

	// Collect all slots for this user across all queues
	var userSlots []models.QueueSlot
	for _, queue := range queues {
		// Apply queue status filter if specified
		if queueStatusFilter != "" && string(queue.Status) != queueStatusFilter {
			continue
		}

		slot, err := u.queueSlotsRepo.GetSlotByQueueAndStudent(ctx, queue.ID, userID)
		if err != nil {
			logger.Errorf(ctx, "failed to get slot for queue %s and user %s: %v", queue.ID, userID, err)
			continue
		}

		// Skip if no slot found for this queue
		if slot == nil {
			continue
		}

		// Apply slot status filter if specified
		if statusFilter != "" && string(slot.Status) != statusFilter {
			continue
		}

		userSlots = append(userSlots, *slot)
	}

	return userSlots, nil
}
