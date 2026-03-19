package gormmodels

import (
	"time"

	"github.com/google/uuid"
)

type QueueSlot struct {
	ID         uuid.UUID       `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	QueueID    uuid.UUID       `gorm:"column:queue_id;type:uuid;not null;uniqueIndex:idx_queue_slots_queue_id_student_id;index:idx_slots_queue_timestamp"`
	StudentID  uuid.UUID       `gorm:"column:student_id;type:uuid;not null;uniqueIndex:idx_queue_slots_queue_id_student_id"`
	Status     QueueSlotStatus `gorm:"column:status;type:queue_slot_status;not null;default:'waiting'"`
	SignedUpAt time.Time       `gorm:"column:signed_up_at;not null;autoCreateTime;index:idx_slots_queue_timestamp"`

	Queue      Queue           `gorm:"foreignKey:QueueID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Student    User            `gorm:"foreignKey:StudentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (QueueSlot) TableName() string {
	return "queue_slots"
}
