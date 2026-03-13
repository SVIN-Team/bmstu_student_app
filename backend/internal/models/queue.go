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
	Id        uuid.UUID
	GroupId   uuid.UUID
	SubjectId uuid.UUID
	LessonId  uuid.UUID
	CreatedBy uuid.UUID
	CreatedAt time.Time
	OpensAt   time.Time
	ClosesAt  *time.Time
	MaxSize   *uint
	Status    QueueStatus
	Slots     []*QueueSlot
}

type SlotStatus string

const (
	SlotStatusWaiting SlotStatus = "waiting"
	SlotStatusPassed  SlotStatus = "passed"
	SlotStatusFailed  SlotStatus = "failed"
	SlotStatusNoShow  SlotStatus = "no_show"
)

type QueueSlot struct {
	Id        uuid.UUID
	QueueId   uuid.UUID
	StudentId uuid.UUID
	Status    SlotStatus
	SignUpAt  time.Time
}
