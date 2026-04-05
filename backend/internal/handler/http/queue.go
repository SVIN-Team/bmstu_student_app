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
	GetByID(ctx context.Context, queueID uuid.UUID) (models.Queue, error)
	GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.Queue, error)
	Update(ctx context.Context, headmanID uuid.UUID, queue models.Queue) error
	Delete(ctx context.Context, headmanID, queueID uuid.UUID) error
}

type QueueHandler struct {
	queueUseCase QueueUseCase
	groupUseCase GroupUseCase
}

func NewQueueHandler(queueUseCase QueueUseCase, groupUseCase GroupUseCase) *QueueHandler {
	return &QueueHandler{
		queueUseCase: queueUseCase,
		groupUseCase: groupUseCase,
	}
}

// GetQueues returns list of queues with optional filters
// GET /queues?group_id=xxx&status=open
func (h *QueueHandler) GetQueues(ctx *gin.Context) {
	groupIDStr := ctx.Query("group_id")
	status := ctx.Query("status")
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(ctx.DefaultQuery("per_page", "20"))

	if perPage > 100 {
		perPage = 100
	}

	// If no group_id provided, use the user's group
	var groupID uuid.UUID
	if groupIDStr == "" {
		userGroupID, exists := ctx.Get("user_group_id")
		if !exists {
			ValidationError(ctx, "group_id is required")
			return
		}
		groupID = userGroupID.(uuid.UUID)
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
		response[i] = h.queueToDTO(queue, false)
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

// GetQueueByID returns a single queue by ID
// GET /queues/:queue_id
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

	SuccessResponse(ctx, http.StatusOK, h.queueToDTO(queue, true))
}

// CreateQueue creates a new queue
// POST /queues
func (h *QueueHandler) CreateQueue(ctx *gin.Context) {
	var req dto.CreateQueueRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
		return
	}

	// Get user info from context
	userID, _ := ctx.Get("user_id")
	userGroupID, _ := ctx.Get("user_group_id")

	// TODO: Use proper CreateQueueParams from usecase
	// For now, returning a placeholder
	ErrorResponse(ctx, http.StatusNotImplemented, "NOT_IMPLEMENTED", "Queue creation not fully implemented yet")

	logger.Infof(ctx, "User %v attempting to create queue for group %v", userID, userGroupID)
}

// UpdateQueue updates an existing queue
// PATCH /queues/:queue_id
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

	SuccessResponse(ctx, http.StatusOK, h.queueToDTO(queue, true))
}

// DeleteQueue deletes a queue
// DELETE /queues/:queue_id
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
func (h *QueueHandler) queueToDTO(queue models.Queue, detailed bool) dto.QueueResponse {
	response := dto.QueueResponse{
		ID:       queue.ID.String(),
		Status:   string(queue.Status),
		OpensAt:  queue.OpensAt,
		ClosesAt: queue.ClosesAt,
		MaxSize:  queue.MaxSize,
		Subject: dto.SubjectResponse{
			ID: queue.SubjectID.String(),
			// Name would need to be fetched from repository
		},
		Group: dto.GroupResponse{
			ID: queue.GroupID.String(),
			// Name would need to be fetched from repository
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
