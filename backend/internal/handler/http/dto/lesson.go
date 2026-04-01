package dto

import (
	"time"

	"github.com/google/uuid"
)

// LessonResponse represents the response body for a lesson
type LessonResponse struct {
	ID       string          `json:"id"`
	Subject  SubjectResponse `json:"subject"`
	Teacher  TeacherResponse `json:"teacher"`
	Room     *RoomResponse   `json:"room,omitempty"`
	Type     string          `json:"type"`
	StartsAt time.Time       `json:"starts_at"`
	EndsAt   time.Time       `json:"ends_at"`
	QueueID  *string         `json:"queue_id,omitempty"`
}

// ImportScheduleRequest represents the request for importing schedule
type ImportScheduleRequest struct {
	GroupID uuid.UUID                  `json:"group_id" binding:"required"`
	Lessons []ScheduleImportRowRequest `json:"lessons" binding:"required"`
}

// ScheduleImportRowRequest represents a single lesson in import file
type ScheduleImportRowRequest struct {
	SubjectName string    `json:"subject_name" binding:"required"`
	TeacherName string    `json:"teacher_name" binding:"required"`
	RoomName    string    `json:"room_name"`
	Type        string    `json:"type" binding:"required"`
	StartsAt    time.Time `json:"starts_at" binding:"required"`
	EndsAt      time.Time `json:"ends_at" binding:"required"`
}

// ImportScheduleResponse represents the response for schedule import
type ImportScheduleResponse struct {
	ImportedCount   int      `json:"imported_count"`
	CreatedSubjects []string `json:"created_subjects,omitempty"`
	CreatedTeachers []string `json:"created_teachers,omitempty"`
	Errors          []string `json:"errors,omitempty"`
}
