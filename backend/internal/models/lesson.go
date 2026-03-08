package models

import (
	"time"

	"github.com/google/uuid"
)

type LessonType string

const (
	LessonTypeLecture LessonType = "lecture"
	LessonTypeSeminar LessonType = "seminar"
	LessonTypeLab     LessonType = "lab"
)

type Lesson struct {
	ID         uuid.UUID
	GroupID    uuid.UUID
	TeacherID  uuid.UUID
	SubjectID  uuid.UUID
	RoomID     uuid.UUID
	LessonType LessonType
	StartsAt   time.Time
	EndsAt     time.Time
}
