package dto

import (
	"time"

	"github.com/google/uuid"
)

// CreateQueueRequest represents the request body for creating a queue
type CreateQueueRequest struct {
	SubjectID uuid.UUID  `json:"subject_id" binding:"required"`
	LessonID  *uuid.UUID `json:"lesson_id"`
	OpensAt   time.Time  `json:"opens_at" binding:"required"`
	ClosesAt  *time.Time `json:"closes_at"`
	MaxSize   *uint32    `json:"max_size"`
}

// UpdateQueueRequest represents the request body for updating a queue
type UpdateQueueRequest struct {
	Status   *string    `json:"status"`
	ClosesAt *time.Time `json:"closes_at"`
	MaxSize  *uint32    `json:"max_size"`
}

// QueueResponse represents the response body for a queue
type QueueResponse struct {
	ID         string               `json:"id"`
	Subject    SubjectResponse      `json:"subject"`
	Group      GroupResponse        `json:"group"`
	LessonID   *string              `json:"lesson_id,omitempty"`
	Status     string               `json:"status"`
	OpensAt    time.Time            `json:"opens_at"`
	ClosesAt   *time.Time           `json:"closes_at,omitempty"`
	MaxSize    *uint32              `json:"max_size,omitempty"`
	SlotsCount *int                 `json:"slots_count,omitempty"`
	CreatedBy  *UserMinimalResponse `json:"created_by,omitempty"`
	CreatedAt  *time.Time           `json:"created_at,omitempty"`
}

// TransferSlotsRequest represents the request body for transferring failed students
type TransferSlotsRequest struct {
	SourceQueueID uuid.UUID `json:"source_queue_id" binding:"required"`
}
