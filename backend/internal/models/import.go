package models

import "time"

type ScheduleImportRow struct {
	GroupName         string
	SubjectName       string
	TeacherLastName   string
	TeacherFirstName  string
	TeacherPatronymic string
	RoomName          string
	LessonType        LessonType
	StartsAt          time.Time
	EndsAt            time.Time
	LineNum           int
}
