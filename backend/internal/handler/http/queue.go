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

type QueueUseCase interface {
    Create(ctx context.Context, params models.CreateQueueParams) (uuid.UUID, error)
    GetByID(ctx context.Context, queueID uuid.UUID) (models.Queue, error)
    GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.Queue, error)
    Update(ctx context.Context, headmanID uuid.UUID, queue models.Queue) error
    Delete(ctx context.Context, headmanID, queueID uuid.UUID) error
}

type SubjectUseCase interface {
    GetByID(ctx context.Context, id uuid.UUID) (models.Subject, error)
}

type QueueHandler struct {
    queueUseCase   QueueUseCase
    groupUseCase   GroupUseCase
    subjectUseCase SubjectUseCase
    userUseCase    UserUseCase
}

func NewQueueHandler(queueUseCase QueueUseCase, groupUseCase GroupUseCase,
    subjectUseCase SubjectUseCase, userUseCase UserUseCase) *QueueHandler {
    return &QueueHandler{
        queueUseCase:   queueUseCase,
        groupUseCase:   groupUseCase,
        subjectUseCase: subjectUseCase,
        userUseCase:    userUseCase,
    }
}

// GetQueues godoc
// @Summary List queues
// @Description Returns queues filtered by group and status with pagination.
// @Tags Queues
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param group_id query string false "Group ID (UUID). If omitted, current user's group is used."
// @Param status query string false "Queue status filter"
// @Param page query int false "Page number" default(1)
// @Param per_page query int false "Items per page" default(20)
// @Success 200 {object} dto.SuccessResponse{data=dto.QueueListResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues [get]
func (h *QueueHandler) GetQueues(ctx *gin.Context) {
    groupIDStr := ctx.Query("group_id")
    status := ctx.Query("status")
    pageStr := ctx.DefaultQuery("page", "1")
    perPageStr := ctx.DefaultQuery("per_page", "20")

    page, err := strconv.Atoi(pageStr)
    if err != nil || page < 1 {
        ValidationError(ctx, "Invalid page: must be a positive integer")
        return
    }

    perPage, err := strconv.Atoi(perPageStr)
    if err != nil || perPage < 1 {
        ValidationError(ctx, "Invalid per_page: must be a positive integer")
        return
    }
    if perPage > 100 {
        perPage = 100
    }

    userId := ctx.MustGet("user_id").(uuid.UUID)

    // If no group_id provided, use the user's group
    var groupID uuid.UUID
    if groupIDStr == "" {
        user, err := h.userUseCase.GetUserByID(ctx, userId)
        if err != nil {
            InternalError(ctx, fmt.Sprintf("Failed to get user group info when group_id is not specified explicitly: %v", err))
            logger.Errorf(ctx, "Failed to get user group info when group_id is not specified explicitly: %v", err)
            return
        }
        groupID = user.GroupID
    } else {
        var err error
        groupID, err = uuid.Parse(groupIDStr)
        if err != nil {
            ValidationError(ctx, "Invalid group_id format")
            return
        }
    }

    queues, err := h.queueUseCase.GetByGroupID(ctx, groupID)
    if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to get queues: %v", err))
        logger.Errorf(ctx, "Failed to get queues: %v", err)
        return
    }

    // Filter by status if provided
    if status != "" {
        var filtered []models.Queue
        for _, q := range queues {
            if string(q.Status) == status {
                filtered = append(filtered, q)
            }
        }
        queues = filtered
    }

    // Simple pagination
    total := len(queues)
    start := (page - 1) * perPage
    end := start + perPage

    if start >= total {
        queues = []models.Queue{}
    } else if end > total {
        queues = queues[start:]
    } else {
        queues = queues[start:end]
    }

    response := make([]dto.QueueResponse, len(queues))
    for i, queue := range queues {
        response[i] = h.queueToDTO(ctx, queue, false)
    }

    SuccessResponse(ctx, http.StatusOK, dto.QueueListResponse{
        Items: response,
        Pagination: dto.PaginationResponse{
            Page:    page,
            PerPage: perPage,
            Total:   total,
        },
    })
}

// GetQueueByID godoc
// @Summary Get queue by ID
// @Description Returns detailed queue information by queue ID.
// @Tags Queues
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Queue ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.QueueResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id} [get]
func (h *QueueHandler) GetQueueByID(ctx *gin.Context) {
    queueIDStr := ctx.Param("queue_id")
    queueID, err := uuid.Parse(queueIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid queue ID format")
        return
    }

    queue, err := h.queueUseCase.GetByID(ctx, queueID)
    if errors.Is(err, errors2.ErrQueueNotFound) {
        NotFoundError(ctx, "Queue not found")
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to get queue: %v", err))
        logger.Errorf(ctx, "Failed to get queue: %v", err)
        return
    }

    SuccessResponse(ctx, http.StatusOK, h.queueToDTO(ctx, queue, true))
}

// CreateQueue godoc
// @Summary Create queue
// @Description Creates a queue for the authenticated user's group.
// @Tags Queues
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param request body dto.CreateQueueRequest true "Create queue request"
// @Success 201 {object} dto.SuccessResponse{data=dto.QueueResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Router /queues [post]
func (h *QueueHandler) CreateQueue(ctx *gin.Context) {
    var req dto.CreateQueueRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
        return
    }

    // Get user info from context
    userID, _ := ctx.Get("user_id")

    // Use proper CreateQueueParams from usecase
    params := models.CreateQueueParams{
        LessonID:           req.LessonID,
        CreatedByUserID:    userID.(uuid.UUID),
        OpensAt:            req.OpensAt,
        ClosesAt:           req.ClosesAt,
        MaxSize:            req.MaxSize,
        TransferFailedFrom: req.TransferFailedFrom,
    }

    queueID, err := h.queueUseCase.Create(ctx, params)
    if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to create queue: %v", err))
        return
    }

    // Get created queue
    queue, err := h.queueUseCase.GetByID(ctx, queueID)
    if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to get created queue: %v", err))
        return
    }

    SuccessResponse(ctx, http.StatusCreated, h.queueToDTO(ctx, queue, true))
}

// UpdateQueue godoc
// @Summary Update queue
// @Description Updates queue status and limits.
// @Tags Queues
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Queue ID"
// @Param request body dto.UpdateQueueRequest true "Update queue request"
// @Success 200 {object} dto.SuccessResponse{data=dto.QueueResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id} [patch]
func (h *QueueHandler) UpdateQueue(ctx *gin.Context) {
    queueIDStr := ctx.Param("queue_id")
    queueID, err := uuid.Parse(queueIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid queue ID format")
        return
    }

    var req dto.UpdateQueueRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
        return
    }

    // Get existing queue
    queue, err := h.queueUseCase.GetByID(ctx, queueID)
    if errors.Is(err, errors2.ErrQueueNotFound) {
        NotFoundError(ctx, "Queue not found")
        return
    }

    // Apply updates
    if req.ClosesAt != nil {
        queue.ClosesAt = req.ClosesAt
    }
    if req.MaxSize != nil {
        queue.MaxSize = req.MaxSize
    }
    if req.Status != nil {
        queue.Status = models.QueueStatus(*req.Status)
    }

    userID := ctx.MustGet("user_id").(uuid.UUID)
    err = h.queueUseCase.Update(ctx, userID, queue)
    if errors.Is(err, errors2.ErrForbidden) {
        ForbiddenError(ctx, "You don't have permission to update this queue")
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to update queue: %v", err))
        return
    }

    SuccessResponse(ctx, http.StatusOK, h.queueToDTO(ctx, queue, true))
}

// DeleteQueue godoc
// @Summary Delete queue
// @Description Deletes a queue by ID.
// @Tags Queues
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Queue ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id} [delete]
func (h *QueueHandler) DeleteQueue(ctx *gin.Context) {
    queueIDStr := ctx.Param("queue_id")
    queueID, err := uuid.Parse(queueIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid queue ID format")
        return
    }

    userID := ctx.MustGet("user_id").(uuid.UUID)
    err = h.queueUseCase.Delete(ctx, userID, queueID)
    if errors.Is(err, errors2.ErrQueueNotFound) {
        NotFoundError(ctx, "Queue not found")
        return
    } else if errors.Is(err, errors2.ErrForbidden) {
        ForbiddenError(ctx, "You don't have permission to delete this queue")
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to delete queue: %v", err))
        return
    }

    ctx.JSON(http.StatusNoContent, nil)
}

// Helper function to convert queue to DTO
func (h *QueueHandler) queueToDTO(ctx context.Context, queue models.Queue, detailed bool) dto.QueueResponse {
    var groupName string
    if h.groupUseCase != nil && queue.GroupID != uuid.Nil {
        if group, err := h.groupUseCase.GetByID(ctx, queue.GroupID); err == nil {
            groupName = group.Name
        } else {
            logger.Errorf(ctx, "failed to fetch group name for group %s: %v", queue.GroupID, err)
        }
    }

    var subjectName string
    if h.subjectUseCase != nil && queue.SubjectID != uuid.Nil {
        if subject, err := h.subjectUseCase.GetByID(ctx, queue.SubjectID); err == nil {
            subjectName = subject.Name
        } else {
            logger.Errorf(ctx, "failed to fetch subject name for subject %s: %v", queue.SubjectID, err)
        }
    }

    response := dto.QueueResponse{
        ID:       queue.ID.String(),
        Status:   string(queue.Status),
        OpensAt:  queue.OpensAt,
        ClosesAt: queue.ClosesAt,
        MaxSize:  queue.MaxSize,
        Subject: dto.SubjectResponse{
            ID:   queue.SubjectID.String(),
            Name: subjectName,
        },
        Group: dto.GroupResponse{
            ID:   queue.GroupID.String(),
            Name: groupName,
        },
    }

    if queue.LessonID != uuid.Nil {
        lessonIDStr := queue.LessonID.String()
        response.LessonID = &lessonIDStr
    }

    if detailed {
        response.CreatedAt = &queue.CreatedAt
        // CreatedBy would need user info from repository
    }

    return response
}
