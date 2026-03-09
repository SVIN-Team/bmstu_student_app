package models

import (
	"time"

	"github.com/google/uuid"
)

type QueueStatus string

const (
	QueueStatusDraft  QueueStatus = "draft"
	QueueStatusOpen   QueueStatus = "open"
	QueueStatusClosed QueueStatus = "closed"
)

type Queue struct {
	ID             uuid.UUID
	GroupID        uuid.UUID
	SubjectID      uuid.UUID
	LessonID       uuid.UUID
	CreatedByUserID uuid.UUID
	CreatedAt      time.Time
	OpensAt        time.Time
	ClosesAt       *time.Time
	MaxSize        *uint32
	Status         QueueStatus
	Slots          []*QueueSlot
}

type SlotStatus string

const (
	SlotStatusWaiting SlotStatus = "waiting"
	SlotStatusPassed  SlotStatus = "passed"
	SlotStatusFailed  SlotStatus = "failed"
	SlotStatusNoShow  SlotStatus = "no_show"
)

type QueueSlot struct {
	ID         uuid.UUID
	QueueID    uuid.UUID
	StudentID  uuid.UUID
	Status     SlotStatus
	SignedUpAt time.Time
}
