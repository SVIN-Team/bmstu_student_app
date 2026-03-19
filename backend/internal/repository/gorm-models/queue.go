package gormmodels

import (
	"time"

	"github.com/google/uuid"
)

type Queue struct {
	ID        uuid.UUID   `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	GroupID   uuid.UUID   `gorm:"column:group_id;type:uuid;not null;index:idx_queues_group_status"`
	SubjectID uuid.UUID   `gorm:"column:subject_id;type:uuid;not null"`
	LessonID  *uuid.UUID  `gorm:"column:lesson_id;type:uuid"`
	CreatedBy uuid.UUID   `gorm:"column:created_by;type:uuid;not null"`
	CreatedAt time.Time   `gorm:"column:created_at;not null;autoCreateTime"`
	OpensAt   time.Time   `gorm:"column:opens_at;not null"`
	ClosesAt  *time.Time  `gorm:"column:closes_at"`
	MaxSize   *int        `gorm:"column:max_size"`
	Status    QueueStatus `gorm:"column:status;type:queue_status;not null;default:'draft';index:idx_queues_group_status"`
	
	Group     Group       `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Subject   Subject     `gorm:"foreignKey:SubjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Lesson    *Lesson     `gorm:"foreignKey:LessonID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Creator   User        `gorm:"foreignKey:CreatedBy;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Slots     []QueueSlot `gorm:"foreignKey:QueueID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (Queue) TableName() string {
	return "queues"
}
