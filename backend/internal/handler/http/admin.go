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

// GetUsers godoc
// @Summary List users (admin)
// @Description Returns paginated list of users.
// @Tags Admin Users
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} dto.SuccessResponse{data=dto.UserListResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users [get]
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
        response[i] = h.userToDTO(ctx, user, true)
    }

    SuccessResponse(ctx, http.StatusOK, dto.UserListResponse{
        Items: response,
        Pagination: dto.PaginationResponse{
            Page:    page,
            PerPage: perPage,
            Total:   total,
        },
    })
}

// GetUserByID godoc
// @Summary Get user by ID (admin)
// @Description Returns user details by ID.
// @Tags Admin Users
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param id path string true "User ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.UserResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users/{id} [get]
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

    SuccessResponse(ctx, http.StatusOK, h.userToDTO(ctx, user, true))
}

// UpdateUser godoc
// @Summary Update user (admin)
// @Description Partially updates user fields.
// @Tags Admin Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param id path string true "User ID"
// @Param request body dto.AdminUpdateUserRequest true "Admin update user request"
// @Success 200 {object} dto.SuccessResponse{data=dto.UserResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users/{id} [patch]
func (h *AdminHandler) UpdateUser(ctx *gin.Context) {
    userIDStr := ctx.Param("id")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid user ID format")
        return
    }

    var req dto.AdminUpdateUserRequest

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

    SuccessResponse(ctx, http.StatusOK, h.userToDTO(ctx, updatedUser, true))
}

// DeleteUser godoc
// @Summary Delete user (admin)
// @Description Deletes user by ID.
// @Tags Admin Users
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param id path string true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/users/{id} [delete]
func (h *AdminHandler) DeleteUser(ctx *gin.Context) {
    userIDStr := ctx.Param("id")
    userID, err := uuid.Parse(userIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid user ID format")
        return
    }

    err = h.adminUseCase.DeleteUser(ctx, userID)
    if errors.Is(err, errors2.ErrUserNotFound) {
        NotFoundError(ctx, "User not found")
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to delete user: %v", err))
        logger.Errorf(ctx, "Failed to delete user: %v", err)
        return
    }

    ctx.JSON(http.StatusNoContent, nil)
}

// ==================== Groups Management ====================

// GetGroups godoc
// @Summary Получить все группы (Админ)
// @Description Вернуть список всех существующих групп
// @Tags Admin Groups
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} dto.SuccessResponse{data=[]dto.GroupResponse}
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/groups [get]
func (h *AdminHandler) GetGroups(ctx *gin.Context) {
    groups, err := h.groupUseCase.GetAll(ctx)
    if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to get groups: %v", err))
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

    SuccessResponse(ctx, http.StatusOK, response)
}

// CreateGroup godoc
// @Summary Создать группу (Админ)
// @Description Создать новую группу по имени
// @Tags Admin Groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param request body dto.AdminUpsertGroupRequest true "Create group request"
// @Success 201 {object} dto.SuccessResponse{data=dto.GroupResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/groups [post]
func (h *AdminHandler) CreateGroup(ctx *gin.Context) {
    var req dto.AdminUpsertGroupRequest

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
        return
    }

    group := models.Group{
        ID:   uuid.New(),
        Name: req.Name,
    }

    id, err := h.groupUseCase.Create(ctx, group)
    if err != nil {
        if errors.Is(err, errors2.ErrGroupAlreadyExists) {
            ValidationError(ctx, fmt.Sprintf("Failed to create group: %v", err))
        } else {
            InternalError(ctx, fmt.Sprintf("Failed to create group: %v", err))
        }
        logger.Errorf(ctx, "Failed to create group: %v", err)
        return
    }

    SuccessResponse(ctx, http.StatusCreated, dto.GroupResponse{
        ID:   id.String(),
        Name: req.Name,
    })
}

// GetGroupByID godoc
// @Summary Получить группу по ID (Админ)
// @Description Получить информацию о группе по ID
// @Tags Admin Groups
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param id path string true "Group ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.GroupResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/groups/{id} [get]
func (h *AdminHandler) GetGroupByID(ctx *gin.Context) {
    groupIDStr := ctx.Param("id")
    groupID, err := uuid.Parse(groupIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid group ID format")
        return
    }

    group, err := h.groupUseCase.GetByID(ctx, groupID)
    if errors.Is(err, errors2.ErrGroupNotFound) {
        NotFoundError(ctx, "Group not found")
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to get group: %v", err))
        logger.Errorf(ctx, "Failed to get group: %v", err)
        return
    }

    SuccessResponse(ctx, http.StatusOK, dto.GroupResponse{
        ID:   group.ID.String(),
        Name: group.Name,
    })
}

// UpdateGroup godoc
// @Summary Обновить группу (Админ)
// @Description Обновить имя группы
// @Tags Admin Groups
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param id path string true "Group ID"
// @Param request body dto.AdminUpsertGroupRequest true "Update group request"
// @Success 200 {object} dto.SuccessResponse{data=dto.GroupResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/groups/{id} [patch]
func (h *AdminHandler) UpdateGroup(ctx *gin.Context) {
    groupIDStr := ctx.Param("id")
    groupID, err := uuid.Parse(groupIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid group ID format")
        return
    }

    var req dto.AdminUpsertGroupRequest

    if err := ctx.ShouldBindJSON(&req); err != nil {
        ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
        return
    }

    group := models.Group{
        ID:   groupID,
        Name: req.Name,
    }

    err = h.groupUseCase.Update(ctx, group)
    if errors.Is(err, errors2.ErrGroupNotFound) {
        NotFoundError(ctx, "Group not found")
        return
    } else if err != nil {
        if errors.Is(err, errors2.ErrUniqueViolationFault) {
            ValidationError(ctx, "Failed to update group: group with that name already exists")
        } else {
            InternalError(ctx, fmt.Sprintf("Failed to update group: %v", err))
        }
        logger.Errorf(ctx, "Failed to update group: %v", err)
        return
    }

    SuccessResponse(ctx, http.StatusOK, dto.GroupResponse{
        ID:   group.ID.String(),
        Name: group.Name,
    })
}

// DeleteGroup godoc
// @Summary Удалить группу (Админ)
// @Description Удалить группу насовсем, если в ней нет участников
// @Tags Admin Groups
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param id path string true "Group ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/groups/{id} [delete]
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
        if errors.Is(err, errors2.ErrGroupHasUsers) {
            ValidationError(ctx, fmt.Sprintf("Failed to delete group: %v", err))
        } else {
            InternalError(ctx, fmt.Sprintf("Failed to delete group: %v", err))
        }
        logger.Errorf(ctx, "Failed to delete group: %v", err)
        return
    }

    ctx.JSON(http.StatusNoContent, nil)
}

// ==================== Helper Functions ====================

func (h *AdminHandler) userToDTO(ctx context.Context, user models.User, detailed bool) dto.UserResponse {
    response := dto.UserResponse{
        ID:        user.ID.String(),
        Email:     user.Email,
        FirstName: user.FirstName,
        LastName:  user.LastName,
        Role:      string(user.Role),
    }

    if user.GroupID != uuid.Nil {
        group, err := h.groupUseCase.GetByID(ctx, user.GroupID)
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
