package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	errors2 "stud_hub/internal/errors"
	"stud_hub/internal/handler/http/dto"
	"stud_hub/internal/handler/middleware"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserUseCase interface {
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) (models.User, error)
	TransferHeadmanRole(ctx context.Context, fromUserID, toUserID uuid.UUID) error
	GetUserSlots(ctx context.Context, userID uuid.UUID, statusFilter, queueStatusFilter string) ([]models.QueueSlot, error)
}

type UserHandler struct {
	userUseCase  UserUseCase
	groupUseCase GroupUseCase
}

func NewUserHandler(userUseCase UserUseCase, groupUseCase GroupUseCase) *UserHandler {
	return &UserHandler{
		userUseCase:  userUseCase,
		groupUseCase: groupUseCase,
	}
}

// GetCurrentUser godoc
// @Summary Get current user
// @Description Returns profile for authenticated user.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} dto.SuccessResponse{data=dto.UserResponse}
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me [get]
func (h *UserHandler) GetCurrentUser(ctx *gin.Context) {
	userID := ctx.MustGet(string(middleware.UserIDContextKey)).(uuid.UUID)

	user, err := h.userUseCase.GetUserByID(ctx, userID)
	if errors.Is(err, errors2.ErrUserNotFound) {
		NotFoundError(ctx, "User not found")
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to get user: %v", err))
		logger.Errorf(ctx, "Failed to get user: %v", err)
		return
	}

	response := h.userToDTO(user, true)
	SuccessResponse(ctx, http.StatusOK, response)
}

// UpdateCurrentUser godoc
// @Summary Update current user
// @Description Updates profile fields for authenticated user.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param request body dto.UpdateUserRequest true "Update user request"
// @Success 200 {object} dto.SuccessResponse{data=dto.UserResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me [patch]
func (h *UserHandler) UpdateCurrentUser(ctx *gin.Context) {
	userID := ctx.MustGet(string(middleware.UserIDContextKey)).(uuid.UUID)

	var req dto.UpdateUserRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Get current user
	user, err := h.userUseCase.GetUserByID(ctx, userID)
	if err != nil {
		InternalError(ctx, "Failed to get user")
		return
	}

	// Check if headman is trying to change group
	if user.Role == models.RoleHeadman && req.GroupID != nil && *req.GroupID != user.GroupID {
		ForbiddenError(ctx, "Передайте роль старосты перед сменой группы")
		return
	}

	// Apply updates
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.GroupID != nil {
		user.GroupID = *req.GroupID
	}

	updatedUser, err := h.userUseCase.UpdateUser(ctx, user)
	if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to update user: %v", err))
		logger.Errorf(ctx, "Failed to update user: %v", err)
		return
	}

	SuccessResponse(ctx, http.StatusOK, h.userToDTO(updatedUser, true))
}

// GetCurrentUserSlots godoc
// @Summary Get current user slots
// @Description Returns authenticated user's slots with optional filters.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param status query string false "Slot status filter"
// @Param queue_status query string false "Queue status filter"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.SlotResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me/slots [get]
func (h *UserHandler) GetCurrentUserSlots(ctx *gin.Context) {
	userID := ctx.MustGet(string(middleware.UserIDContextKey)).(uuid.UUID)

	statusFilter := ctx.Query("status")
	queueStatusFilter := ctx.Query("queue_status")

	slots, err := h.userUseCase.GetUserSlots(ctx, userID, statusFilter, queueStatusFilter)
	if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to get user slots: %v", err))
		logger.Errorf(ctx, "Failed to get user slots: %v", err)
		return
	}

	response := make([]dto.SlotResponse, len(slots))
	for i, slot := range slots {
		response[i] = dto.SlotResponse{
			ID:         slot.ID.String(),
			Status:     string(slot.Status),
			SignedUpAt: slot.SignedUpAt,
			Queue: &dto.QueueMinimalInfo{
				ID: slot.QueueID.String(),
				// Subject and Status would need to be fetched from repository
			},
		}
	}

	SuccessResponse(ctx, http.StatusOK, response)
}

// TransferHeadmanRole godoc
// @Summary Transfer headman role
// @Description Transfers headman role to another user in the same group.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param request body dto.TransferHeadmanRoleRequest true "Transfer headman role request"
// @Success 200 {object} dto.SuccessResponse{data=dto.TransferHeadmanRoleResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /users/me/headman-role [put]
func (h *UserHandler) TransferHeadmanRole(ctx *gin.Context) {
	userID := ctx.MustGet(string(middleware.UserIDContextKey)).(uuid.UUID)

	var req dto.TransferHeadmanRoleRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Check if current user is headman
	user, err := h.userUseCase.GetUserByID(ctx, userID)
	if err != nil {
		InternalError(ctx, "Failed to get user")
		return
	}

	if user.Role != models.RoleHeadman {
		ForbiddenError(ctx, "Only headman can transfer the role")
		return
	}

	err = h.userUseCase.TransferHeadmanRole(ctx, userID, req.ToUserID)
	if errors.Is(err, errors2.ErrUserNotFound) {
		NotFoundError(ctx, "Target user not found")
		return
	} else if errors.Is(err, errors2.ErrForbidden) {
		ForbiddenError(ctx, "Target user must be in the same group")
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to transfer role: %v", err))
		logger.Errorf(ctx, "Failed to transfer role: %v", err)
		return
	}

	SuccessResponse(ctx, http.StatusOK, dto.TransferHeadmanRoleResponse{
		PreviousHeadmanID: userID.String(),
		NewHeadmanID:      req.ToUserID.String(),
		GroupID:           user.GroupID.String(),
	})
}

// Helper function to convert user to DTO
func (h *UserHandler) userToDTO(user models.User, detailed bool) dto.UserResponse {
	response := dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
	}

	if user.GroupID != uuid.Nil {
		// Fetch group info
		group, err := h.groupUseCase.GetByID(context.Background(), user.GroupID)
		if err == nil {
			response.Group = &dto.GroupResponse{
				ID:   group.ID.String(),
				Name: group.Name,
			}
		}
	}

	if detailed {
		response.IsBlocked = &user.IsBlocked
		response.CreatedAt = &user.CreatedAt
	}

	return response
}
