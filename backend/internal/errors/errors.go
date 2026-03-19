package errors

import "errors"

// Auth
var (
	ErrInvalidCredentials  = errors.New("invalid email or password")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInternalServer      = errors.New("internal server error")
)

// Schedule
var (

	ErrLessonNotFound    = errors.New("lesson not found")
	ErrSubjectNotFound   = errors.New("subject not found")
	ErrGroupNotFound     = errors.New("group not found")
	ErrInvalidDateRange  = errors.New("invalid date range")
	ErrLessonOverlap     = errors.New("lesson time overlaps with existing lesson")
	ErrInvalidLessonTime = errors.New("lesson end time must be after start time")
)

// Group
var (
	ErrGroupAlreadyExists = errors.New("group already exists")
	ErrGroupHasUsers      = errors.New("group has users")
	ErrForbidden          = errors.New("insufficient privileges to perform this action")
)

// Queue
var (
	ErrQueueNotFound    = errors.New("queue not found")
	ErrQueueNotOpen     = errors.New("queue is not open")
	ErrQueueClosed      = errors.New("queue is closed")
	ErrQueueFull        = errors.New("queue is full")
	ErrQueueAlreadyOpen = errors.New("queue is already open")
	ErrInvalidQueueTime = errors.New("invalid queue time range")
	ErrAlreadyInQueue   = errors.New("already signed up for this queue")
	ErrNotInQueue       = errors.New("not signed up for this queue")
	ErrNotInQueueGroup  = errors.New("not a member of queue's group")
	ErrSlotNotFound     = errors.New("slot not found")
)

// Admin errors
var (
	ErrUserNotFound        = errors.New("user not found")
	ErrNotGroupMember      = errors.New("user is not a member of any group")
	ErrImportFailed        = errors.New("schedule import failed")
	ErrInvalidImportFormat = errors.New("invalid import file format")
)

var (
	ErrTeacherNotFound   = errors.New("teacher not found")
	ErrClassroomNotFound = errors.New("classroom not found")
)
