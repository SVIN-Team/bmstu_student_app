package gormmodels

// Enum types mirror PostgreSQL enums defined in deploy/postgresql/001-init.sql.
type (
	UserRole        string
	LessonType      string
	QueueStatus     string
	QueueSlotStatus string
)

const (
	UserRoleStudent UserRole = "student"
	UserRoleHeadman UserRole = "headman"
	UserRoleAdmin   UserRole = "admin"

	LessonTypeLecture LessonType = "lecture"
	LessonTypeLab     LessonType = "lab"
	LessonTypeSeminar LessonType = "seminar"

	QueueStatusDraft    QueueStatus = "draft"
	QueueStatusOpen     QueueStatus = "open"
	QueueStatusClosed   QueueStatus = "closed"
	QueueStatusArchived QueueStatus = "archived"

	QueueSlotStatusWaiting QueueSlotStatus = "waiting"
	QueueSlotStatusPassed  QueueSlotStatus = "passed"
	QueueSlotStatusFailed  QueueSlotStatus = "failed"
	QueueSlotStatusNoShow  QueueSlotStatus = "no_show"
)
