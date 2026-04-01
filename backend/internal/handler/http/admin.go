package http

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	errors2 "stud_hub/internal/errors"
	"stud_hub/internal/handler/http/dto"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AdminUseCase interface {
	// User management
	GetAllUsers(ctx context.Context, page, perPage int) ([]models.User, int, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error)
	UpdateUser(ctx context.Context, user models.User) (models.User, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
	BlockUser(ctx context.Context, id uuid.UUID, block bool) error
	ChangeUserRole(ctx context.Context, id uuid.UUID, role models.RoleType) error
}

type AdminHandler struct {
	adminUseCase AdminUseCase
	groupUseCase GroupUseCase
}

func NewAdminHandler(
	adminUseCase AdminUseCase,
	groupUseCase GroupUseCase,
) *AdminHandler {
	return &AdminHandler{
		adminUseCase: adminUseCase,
		groupUseCase: groupUseCase,
	}
}

// ==================== User Management ====================

// GetUsers returns list of all users (admin only)
// GET /admin/users?page=1&per_page=20
func (h *AdminHandler) GetUsers(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))

	if perPage > 100 {
		perPage = 100
	}

	users, total, err := h.adminUseCase.GetAllUsers(ctx, page, perPage)
	if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to get users: %v", err))
		logger.Errorf(ctx, "Failed to get users: %v", err)
		return
	}

	response := make([]dto.UserResponse, len(users))
	for i, user := range users {
		response[i] = h.userToDTO(user, true)
	}

	SuccessResponse(ctx, http.StatusOK, gin.H{
		"items": response,
		"pagination": dto.PaginationResponse{
			Page:    page,
			PerPage: perPage,
			Total:   total,
		},
	})
}

// GetUserByID returns a specific user by ID (admin only)
// GET /admin/users/:id
func (h *AdminHandler) GetUserByID(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ValidationError(ctx, "Invalid user ID format")
		return
	}

	user, err := h.adminUseCase.GetUserByID(ctx, userID)
	if errors.Is(err, errors2.ErrUserNotFound) {
		NotFoundError(ctx, "User not found")
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to get user: %v", err))
		logger.Errorf(ctx, "Failed to get user: %v", err)
		return
	}

	SuccessResponse(ctx, http.StatusOK, h.userToDTO(user, true))
}

// UpdateUser updates user information (admin only)
// PATCH /admin/users/:id
func (h *AdminHandler) UpdateUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ValidationError(ctx, "Invalid user ID format")
		return
	}

	var req struct {
		Role      *string `json:"role"`
		IsBlocked *bool   `json:"is_blocked"`
		GroupID   *string `json:"group_id"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	user, err := h.adminUseCase.GetUserByID(ctx, userID)
	if errors.Is(err, errors2.ErrUserNotFound) {
		NotFoundError(ctx, "User not found")
		return
	}

	// Apply updates
	if req.Role != nil {
		err = h.adminUseCase.ChangeUserRole(ctx, userID, models.RoleType(*req.Role))
		if err != nil {
			InternalError(ctx, fmt.Sprintf("Failed to change role: %v", err))
			return
		}
		user.Role = models.RoleType(*req.Role)
	}

	if req.IsBlocked != nil {
		err = h.adminUseCase.BlockUser(ctx, userID, *req.IsBlocked)
		if err != nil {
			InternalError(ctx, fmt.Sprintf("Failed to block/unblock user: %v", err))
			return
		}
		user.IsBlocked = *req.IsBlocked
	}

	if req.GroupID != nil {
		groupID, err := uuid.Parse(*req.GroupID)
		if err != nil {
			ValidationError(ctx, "Invalid group ID format")
			return
		}
		user.GroupID = groupID
	}

	updatedUser, err := h.adminUseCase.UpdateUser(ctx, user)
	if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to update user: %v", err))
		logger.Errorf(ctx, "Failed to update user: %v", err)
		return
	}

	SuccessResponse(ctx, http.StatusOK, h.userToDTO(updatedUser, true))
}

// DeleteUser deletes a user (admin only)
// DELETE /admin/users/:id
func (h *AdminHandler) DeleteUser(ctx *gin.Context) {
	userIDStr := ctx.Param("id")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid user ID format",
			},
		})
		return
	}

	err = h.adminUseCase.DeleteUser(ctx, userID)
	if errors.Is(err, errors2.ErrUserNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "User not found",
			},
		})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to delete user: %v", err),
			},
		})
		logger.Errorf(ctx, "Failed to delete user: %v", err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// ==================== Groups Management ====================

// GetGroups returns all groups
// GET /admin/groups
func (h *AdminHandler) GetGroups(ctx *gin.Context) {
	groups, err := h.groupUseCase.GetAll(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to get groups: %v", err),
			},
		})
		logger.Errorf(ctx, "Failed to get groups: %v", err)
		return
	}

	response := make([]dto.GroupResponse, len(groups))
	for i, group := range groups {
		response[i] = dto.GroupResponse{
			ID:   group.ID.String(),
			Name: group.Name,
		}
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// CreateGroup creates a new group
// POST /admin/groups
func (h *AdminHandler) CreateGroup(ctx *gin.Context) {
	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": fmt.Sprintf("Invalid request body: %v", err),
			},
		})
		return
	}

	group := models.Group{
		ID:   uuid.New(),
		Name: req.Name,
	}

	id, err := h.groupUseCase.Create(ctx, group)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to create group: %v", err),
			},
		})
		logger.Errorf(ctx, "Failed to create group: %v", err)
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data": dto.GroupResponse{
			ID:   id.String(),
			Name: req.Name,
		},
	})
}

// GetGroupByID returns a specific group
// GET /admin/groups/:id
func (h *AdminHandler) GetGroupByID(ctx *gin.Context) {
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid group ID format",
			},
		})
		return
	}

	group, err := h.groupUseCase.GetByID(ctx, groupID)
	if errors.Is(err, errors2.ErrGroupNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Group not found",
			},
		})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to get group: %v", err),
			},
		})
		logger.Errorf(ctx, "Failed to get group: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": dto.GroupResponse{
			ID:   group.ID.String(),
			Name: group.Name,
		},
	})
}

// UpdateGroup updates a group
// PATCH /admin/groups/:id
func (h *AdminHandler) UpdateGroup(ctx *gin.Context) {
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid group ID format",
			},
		})
		return
	}

	var req struct {
		Name string `json:"name" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": fmt.Sprintf("Invalid request body: %v", err),
			},
		})
		return
	}

	group := models.Group{
		ID:   groupID,
		Name: req.Name,
	}

	err = h.groupUseCase.Update(ctx, group)
	if errors.Is(err, errors2.ErrGroupNotFound) {
		ctx.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_FOUND",
				"message": "Group not found",
			},
		})
		return
	} else if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": fmt.Sprintf("Failed to update group: %v", err),
			},
		})
		logger.Errorf(ctx, "Failed to update group: %v", err)
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": dto.GroupResponse{
			ID:   group.ID.String(),
			Name: group.Name,
		},
	})
}

// DeleteGroup deletes a group
// DELETE /admin/groups/:id
func (h *AdminHandler) DeleteGroup(ctx *gin.Context) {
	groupIDStr := ctx.Param("id")
	groupID, err := uuid.Parse(groupIDStr)
	if err != nil {
		ValidationError(ctx, "Invalid group ID format")
		return
	}

	err = h.groupUseCase.Delete(ctx, groupID)
	if errors.Is(err, errors2.ErrGroupNotFound) {
		NotFoundError(ctx, "Group not found")
		return
	} else if err != nil {
		InternalError(ctx, fmt.Sprintf("Failed to delete group: %v", err))
		logger.Errorf(ctx, "Failed to delete group: %v", err)
		return
	}

	ctx.JSON(http.StatusNoContent, nil)
}

// ==================== Helper Functions ====================

func (h *AdminHandler) userToDTO(user models.User, detailed bool) dto.UserResponse {
	response := dto.UserResponse{
		ID:        user.ID.String(),
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Role:      string(user.Role),
	}

	if user.GroupID != uuid.Nil {
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
