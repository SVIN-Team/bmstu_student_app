package gormmodels

import (
	"time"

	"github.com/google/uuid"
)

type Lesson struct {
	ID        uuid.UUID  `gorm:"column:id;type:uuid;default:gen_random_uuid();primaryKey"`
	GroupID   uuid.UUID  `gorm:"column:group_id;type:uuid;not null;index:idx_lessons_group_date"`
	SubjectID uuid.UUID  `gorm:"column:subject_id;type:uuid;not null"`
	TeacherID uuid.UUID  `gorm:"column:teacher_id;type:uuid;not null"`
	RoomID    *uuid.UUID `gorm:"column:room_id;type:uuid"`
	Type      LessonType `gorm:"column:type;type:lesson_type;not null"`
	StartsAt  time.Time  `gorm:"column:starts_at;not null;index:idx_lessons_group_date"`
	EndsAt    time.Time  `gorm:"column:ends_at;not null"`

	Group     Group      `gorm:"foreignKey:GroupID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Subject   Subject    `gorm:"foreignKey:SubjectID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Teacher   Teacher    `gorm:"foreignKey:TeacherID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;"`
	Room      *Room      `gorm:"foreignKey:RoomID;references:ID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
	Queues    []Queue    `gorm:"foreignKey:LessonID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;"`
}

func (Lesson) TableName() string {
	return "lessons"
}
