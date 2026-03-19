package gormmodels

import (
	"time"

	"github.com/google/uuid"
)

type QueueSlot struct {
	ID         uuid.UUID       `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	QueueID    uuid.UUID       `gorm:"column:queue_id;type:uuid;not null;unique;uniqueIndex:queue_slots_queue_id_key;index:idx_slots_queue_timestamp"`
	StudentID  uuid.UUID       `gorm:"column:student_id;type:uuid;not null;unique;uniqueIndex:queue_slots_student_id_key"`
	Status     QueueSlotStatus `gorm:"column:status;type:queue_slot_status;not null;default:'waiting'"`
	SignedUpAt time.Time       `gorm:"column:signed_up_at;not null;autoCreateTime;index:idx_slots_queue_timestamp"`
	//Version    int             `gorm:"column:version;not null;default:1"`

	Queue      Queue           `gorm:"foreignKey:QueueID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Student    User            `gorm:"foreignKey:StudentID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (QueueSlot) TableName() string {
	return "queue_slots"
}
