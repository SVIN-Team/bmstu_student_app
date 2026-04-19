package http

import (
    "context"
    "errors"
    "fmt"
    "net/http"

    errors2 "stud_hub/internal/errors"
    "stud_hub/internal/handler/http/dto"
    "stud_hub/internal/models"
    "stud_hub/util/logger"

    "github.com/gin-gonic/gin"
    "github.com/google/uuid"
)

type QueueSlotsUseCase interface {
    GetSlotsByQueueID(ctx context.Context, queueID uuid.UUID) ([]models.QueueSlot, error)
    GetSlotByID(ctx context.Context, slotID uuid.UUID) (models.QueueSlot, error)
    SignUp(ctx context.Context, studentID, queueID uuid.UUID) (*models.QueueSlot, error)
    CancelSignUp(ctx context.Context, studentID, queueID uuid.UUID) error
    UpdateSlotStatus(ctx context.Context, headmanID, slotID uuid.UUID, status models.SlotStatus) error
    TransferFailed(ctx context.Context, headmanID, fromQueueID, toQueueID uuid.UUID) (int, error)
}

type QueueSlotsHandler struct {
    queueSlotsUseCase QueueSlotsUseCase
    queueUseCase      QueueUseCase
}

func NewQueueSlotsHandler(queueSlotsUseCase QueueSlotsUseCase, queueUseCase QueueUseCase) *QueueSlotsHandler {
    return &QueueSlotsHandler{
        queueSlotsUseCase: queueSlotsUseCase,
        queueUseCase:      queueUseCase,
    }
}

// GetQueueSlots godoc
// @Summary List queue slots
// @Description Returns all slots in a queue.
// @Tags Queue Slots
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Queue ID"
// @Success 200 {object} dto.SuccessResponse{data=[]dto.SlotResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id}/slots [get]
func (h *QueueSlotsHandler) GetQueueSlots(ctx *gin.Context) {
    queueIDStr := ctx.Param("queue_id")
    queueID, err := uuid.Parse(queueIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid queue ID format")
        return
    }

    // Check if queue exists
    _, err = h.queueUseCase.GetByID(ctx, queueID)
    if errors.Is(err, errors2.ErrQueueNotFound) {
        NotFoundError(ctx, "Queue not found")
        return
    }

    slots, err := h.queueSlotsUseCase.GetSlotsByQueueID(ctx, queueID)
    if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to get slots: %v", err))
        logger.Errorf(ctx, "Failed to get slots: %v", err)
        return
    }

    response := make([]dto.SlotResponse, len(slots))
    for i, slot := range slots {
        response[i] = h.slotToDTO(slot, false)
    }

    SuccessResponse(ctx, http.StatusOK, response)
}

// GetQueueSlot godoc
// @Summary Get queue slot by ID
// @Description Returns slot details by slot ID.
// @Tags Queue Slots
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Queue ID"
// @Param slot_id path string true "Slot ID"
// @Success 200 {object} dto.SuccessResponse{data=dto.SlotResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id}/slots/{slot_id} [get]
func (h *QueueSlotsHandler) GetQueueSlot(ctx *gin.Context) {
    slotIDStr := ctx.Param("slot_id")
    slotID, err := uuid.Parse(slotIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid slot ID format")
        return
    }

    slot, err := h.queueSlotsUseCase.GetSlotByID(ctx, slotID)
    if errors.Is(err, errors2.ErrSlotNotFound) {
        NotFoundError(ctx, "Slot not found")
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to get slot: %v", err))
        logger.Errorf(ctx, "Failed to get slot: %v", err)
        return
    }

    SuccessResponse(ctx, http.StatusOK, h.slotToDTO(slot, true))
}

// SignUpForQueue godoc
// @Summary Sign up for queue
// @Description Signs up authenticated user for queue and returns created slot data.
// @Tags Queue Slots
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Queue ID"
// @Success 201 {object} dto.SuccessResponse{data=dto.SlotResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id}/slots [post]
func (h *QueueSlotsHandler) SignUpForQueue(ctx *gin.Context) {
    queueIDStr := ctx.Param("queue_id")
    queueID, err := uuid.Parse(queueIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid queue ID format")
        return
    }

    userID := ctx.MustGet("user_id").(uuid.UUID)

    newSlot, err := h.queueSlotsUseCase.SignUp(ctx, userID, queueID)
    if errors.Is(err, errors2.ErrQueueNotFound) {
        NotFoundError(ctx, "Queue not found")
        return
    } else if errors.Is(err, errors2.ErrQueueNotOpen) {
        QueueClosedError(ctx)
        return
    } else if errors.Is(err, errors2.ErrQueueFull) {
        QueueFullError(ctx)
        return
    } else if errors.Is(err, errors2.ErrAlreadyInQueue) {
        AlreadyInQueueError(ctx)
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to sign up: %v", err))
        logger.Errorf(ctx, "Failed to sign up: %v", err)
        return
    }

    SuccessResponse(ctx, http.StatusCreated, h.slotToDTO(*newSlot, true))
}

// UpdateQueueSlot godoc
// @Summary Update queue slot
// @Description Updates queue slot status.
// @Tags Queue Slots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Queue ID"
// @Param slot_id path string true "Slot ID"
// @Param request body dto.UpdateSlotRequest true "Update slot request"
// @Success 200 {object} dto.SuccessResponse{data=dto.SlotResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id}/slots/{slot_id} [patch]
func (h *QueueSlotsHandler) UpdateQueueSlot(ctx *gin.Context) {
    slotIDStr := ctx.Param("slot_id")
    slotID, err := uuid.Parse(slotIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid slot ID format")
        return
    }

    var req dto.UpdateSlotRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
        return
    }

    userID := ctx.MustGet("user_id").(uuid.UUID)
    status := models.SlotStatus(req.Status)

    err = h.queueSlotsUseCase.UpdateSlotStatus(ctx, userID, slotID, status)
    if errors.Is(err, errors2.ErrSlotNotFound) {
        NotFoundError(ctx, "Slot not found")
        return
    } else if errors.Is(err, errors2.ErrForbidden) {
        ForbiddenError(ctx, "You don't have permission to update this slot")
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to update slot: %v", err))
        logger.Errorf(ctx, "Failed to update slot: %v", err)
        return
    }

    // Get updated slot
    slot, _ := h.queueSlotsUseCase.GetSlotByID(ctx, slotID)
    SuccessResponse(ctx, http.StatusOK, h.slotToDTO(slot, true))
}

// CancelQueueSlot godoc
// @Summary Cancel queue slot
// @Description Cancels authenticated user's queue sign-up.
// @Tags Queue Slots
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Queue ID"
// @Param slot_id path string true "Slot ID"
// @Success 204 "No Content"
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id}/slots/{slot_id} [delete]
func (h *QueueSlotsHandler) CancelQueueSlot(ctx *gin.Context) {
    queueIDStr := ctx.Param("queue_id")
    queueID, err := uuid.Parse(queueIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid queue ID format")
        return
    }

    userID := ctx.MustGet("user_id").(uuid.UUID)

    err = h.queueSlotsUseCase.CancelSignUp(ctx, userID, queueID)
    if errors.Is(err, errors2.ErrQueueNotFound) {
        NotFoundError(ctx, "Queue not found")
        return
    } else if errors.Is(err, errors2.ErrNotInQueue) {
        NotFoundError(ctx, "You are not signed up for this queue")
        return
    } else if errors.Is(err, errors2.ErrQueueClosed) {
        QueueClosedError(ctx)
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to cancel sign-up: %v", err))
        logger.Errorf(ctx, "Failed to cancel sign-up: %v", err)
        return
    }

    ctx.JSON(http.StatusNoContent, nil)
}

// TransferFailedSlots godoc
// @Summary Transfer failed slots
// @Description Transfers failed students from source queue to target queue.
// @Tags Queue Slots
// @Accept json
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Param queue_id path string true "Target queue ID"
// @Param request body dto.TransferSlotsRequest true "Transfer slots request"
// @Success 201 {object} dto.SuccessResponse{data=dto.TransferSlotsResponse}
// @Failure 400 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /queues/{queue_id}/transfers [post]
func (h *QueueSlotsHandler) TransferFailedSlots(ctx *gin.Context) {
    queueIDStr := ctx.Param("queue_id")
    toQueueID, err := uuid.Parse(queueIDStr)
    if err != nil {
        ValidationError(ctx, "Invalid queue ID format")
        return
    }

    var req dto.TransferSlotsRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        ValidationError(ctx, fmt.Sprintf("Invalid request body: %v", err))
        return
    }

    userID := ctx.MustGet("user_id").(uuid.UUID)

    count, err := h.queueSlotsUseCase.TransferFailed(ctx, userID, req.SourceQueueID, toQueueID)
    if errors.Is(err, errors2.ErrQueueNotFound) {
        NotFoundError(ctx, "Queue not found")
        return
    } else if errors.Is(err, errors2.ErrForbidden) {
        ForbiddenError(ctx, "You don't have permission to transfer slots")
        return
    } else if err != nil {
        InternalError(ctx, fmt.Sprintf("Failed to transfer slots: %v", err))
        logger.Errorf(ctx, "Failed to transfer slots: %v", err)
        return
    }

    // Get transferred slots
    slots, _ := h.queueSlotsUseCase.GetSlotsByQueueID(ctx, toQueueID)
    response := make([]dto.SlotResponse, 0, len(slots))
    for _, slot := range slots {
        response = append(response, h.slotToDTO(slot, false))
    }

    SuccessResponse(ctx, http.StatusCreated, dto.TransferSlotsResponse{
        TransferredCount: count,
        Slots:            response,
    })
}

// Helper function to convert slot to DTO
func (h *QueueSlotsHandler) slotToDTO(slot models.QueueSlot, detailed bool) dto.SlotResponse {
    response := dto.SlotResponse{
        ID: slot.ID.String(),
        Student: dto.UserMinimalResponse{
            ID: slot.StudentID.String(),
            // FirstName and LastName would need to be fetched from repository
        },
        Status:     string(slot.Status),
        SignedUpAt: slot.SignedUpAt,
    }

    if slot.Position > 0 {
        response.Position = &slot.Position
    }

    if detailed {
        queueIDStr := slot.QueueID.String()
        response.QueueID = &queueIDStr
    }

    return response
}
