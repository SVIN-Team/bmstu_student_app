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
	Id         uuid.UUID
	GroupId    uuid.UUID
	TeacherId  uuid.UUID
	SubjectId  uuid.UUID
	RoomId     uuid.UUID
	LessonType LessonType
	StartsAt   time.Time
	EndsAt     time.Time
}
