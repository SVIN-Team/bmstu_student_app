package dto

import (
	"time"
)

// SlotResponse represents the response body for a slot
type SlotResponse struct {
	ID         string              `json:"id"`
	QueueID    *string             `json:"queue_id,omitempty"`
	Student    UserMinimalResponse `json:"student"`
	Status     string              `json:"status"`
	SignedUpAt time.Time           `json:"signed_up_at"`
	Position   *int                `json:"position,omitempty"`
	Queue      *QueueMinimalInfo   `json:"queue,omitempty"`
}

// QueueMinimalInfo represents minimal queue information
type QueueMinimalInfo struct {
	ID      string          `json:"id"`
	Subject SubjectResponse `json:"subject"`
	Status  string          `json:"status"`
}

// UpdateSlotRequest represents the request body for updating a slot status
type UpdateSlotRequest struct {
	Status string `json:"status" binding:"required"`
}

// TransferSlotsResponse represents the response for slot transfers
type TransferSlotsResponse struct {
	TransferredCount int            `json:"transferred_count"`
	Slots            []SlotResponse `json:"slots"`
}
