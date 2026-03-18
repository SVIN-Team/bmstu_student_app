// internal/usecase/schedule_usecase.go
package usecase

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
)

type LessonRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Lesson, error)
	GetByGroupAndDateRange(ctx context.Context, groupID uuid.UUID, from, to time.Time) ([]models.Lesson, error)
	Create(ctx context.Context, lesson models.Lesson) (uuid.UUID, error)
	Update(ctx context.Context, lesson models.Lesson) error
	Delete(ctx context.Context, id uuid.UUID) error
	CheckOverlap(ctx context.Context, groupID uuid.UUID, startsAt, endsAt time.Time, excludeID *uuid.UUID) (bool, error)
	BulkCreate(ctx context.Context, lessons []models.Lesson) (int, error)
	DeleteByGroupAndDateRange(ctx context.Context, groupID uuid.UUID, from, to time.Time) (int, error)
}

type SubjectRepositoryForSchedule interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Subject, error)
	GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error)
}

type GroupRepositoryForSchedule interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Group, error)
	GetByName(ctx context.Context, name string) (models.Group, error)
}

type TeacherRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Teacher, error)
	GetOrCreateByFullName(ctx context.Context, lastName, firstName, patronymic string) (uuid.UUID, error)
}

type ClassroomRepositoryForSchedule interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Classroom, error)
	GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error)
}

type QueueRepositoryForSchedule interface {
	GetByLessonID(ctx context.Context, lessonID uuid.UUID) (*models.Queue, error)
}

type ScheduleUseCase struct {
	lessonRepo    LessonRepository
	subjectRepo   SubjectRepositoryForSchedule
	groupRepo     GroupRepositoryForSchedule
	teacherRepo   TeacherRepository
	classroomRepo ClassroomRepositoryForSchedule
	queueRepo     QueueRepositoryForSchedule
}

func NewScheduleUseCase(
	lessonRepo LessonRepository,
	subjectRepo SubjectRepositoryForSchedule,
	groupRepo GroupRepositoryForSchedule,
	teacherRepo TeacherRepository,
	classroomRepo ClassroomRepositoryForSchedule,
	queueRepo QueueRepositoryForSchedule,
) *ScheduleUseCase {
	return &ScheduleUseCase{
		lessonRepo:    lessonRepo,
		subjectRepo:   subjectRepo,
		groupRepo:     groupRepo,
		teacherRepo:   teacherRepo,
		classroomRepo: classroomRepo,
		queueRepo:     queueRepo,
	}
}

// ==================== Чтение (для всех) ====================

func (s *ScheduleUseCase) GetSchedule(ctx context.Context, groupID uuid.UUID, from, to time.Time) ([]models.LessonDetails, error) {
	if from.After(to) {
		return nil, apperrors.ErrInvalidDateRange
	}

	maxDuration := 30 * 24 * time.Hour
	if to.Sub(from) > maxDuration {
		to = from.Add(maxDuration)
	}

	if _, err := s.groupRepo.GetByID(ctx, groupID); err != nil {
		return nil, apperrors.ErrGroupNotFound
	}

	lessons, err := s.lessonRepo.GetByGroupAndDateRange(ctx, groupID, from, to)
	if err != nil {
		logger.Errorf(ctx, "failed to get lessons: %v", err)
		return nil, apperrors.ErrInternalServer
	}

	result := make([]models.LessonDetails, 0, len(lessons))
	for _, lesson := range lessons {
		details := s.enrichLessonDetails(ctx, lesson)
		result = append(result, details)
	}

	return result, nil
}

func (s *ScheduleUseCase) GetWeekSchedule(ctx context.Context, groupID uuid.UUID, date time.Time) ([]models.LessonDetails, error) {
	weekStart := getWeekStart(date)
	weekEnd := weekStart.AddDate(0, 0, 7)
	return s.GetSchedule(ctx, groupID, weekStart, weekEnd)
}

func (s *ScheduleUseCase) GetDaySchedule(ctx context.Context, groupID uuid.UUID, date time.Time) ([]models.LessonDetails, error) {
	dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)
	return s.GetSchedule(ctx, groupID, dayStart, dayEnd)
}

func (s *ScheduleUseCase) GetLessonByID(ctx context.Context, lessonID uuid.UUID) (models.LessonDetails, error) {
	lesson, err := s.lessonRepo.GetByID(ctx, lessonID)
	if err != nil {
		return models.LessonDetails{}, apperrors.ErrLessonNotFound
	}
	return s.enrichLessonDetails(ctx, lesson), nil
}

// ==================== CRUD занятий (админ) ====================

func (s *ScheduleUseCase) CreateLesson(ctx context.Context, lesson models.Lesson) (uuid.UUID, error) {
	if err := s.validateLesson(ctx, lesson); err != nil {
		return uuid.Nil, err
	}

	hasOverlap, err := s.lessonRepo.CheckOverlap(ctx, lesson.GroupID, lesson.StartsAt, lesson.EndsAt, nil)
	if err != nil {
		logger.Errorf(ctx, "failed to check overlap: %v", err)
		return uuid.Nil, apperrors.ErrInternalServer
	}
	if hasOverlap {
		return uuid.Nil, apperrors.ErrLessonOverlap
	}

	lesson.ID = uuid.New()

	id, err := s.lessonRepo.Create(ctx, lesson)
	if err != nil {
		logger.Errorf(ctx, "failed to create lesson: %v", err)
		return uuid.Nil, apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "lesson created: %s", id)
	return id, nil
}

func (s *ScheduleUseCase) UpdateLesson(ctx context.Context, lesson models.Lesson) error {
	if _, err := s.lessonRepo.GetByID(ctx, lesson.ID); err != nil {
		return apperrors.ErrLessonNotFound
	}

	if err := s.validateLesson(ctx, lesson); err != nil {
		return err
	}

	hasOverlap, err := s.lessonRepo.CheckOverlap(ctx, lesson.GroupID, lesson.StartsAt, lesson.EndsAt, &lesson.ID)
	if err != nil {
		logger.Errorf(ctx, "failed to check overlap: %v", err)
		return apperrors.ErrInternalServer
	}
	if hasOverlap {
		return apperrors.ErrLessonOverlap
	}

	if err := s.lessonRepo.Update(ctx, lesson); err != nil {
		logger.Errorf(ctx, "failed to update lesson: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "lesson updated: %s", lesson.ID)
	return nil
}

func (s *ScheduleUseCase) DeleteLesson(ctx context.Context, lessonID uuid.UUID) error {
	if _, err := s.lessonRepo.GetByID(ctx, lessonID); err != nil {
		return apperrors.ErrLessonNotFound
	}

	if err := s.lessonRepo.Delete(ctx, lessonID); err != nil {
		logger.Errorf(ctx, "failed to delete lesson: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "lesson deleted: %s", lessonID)
	return nil
}

// ==================== Импорт (админ) ====================

type ImportResult struct {
	TotalRows    int
	SuccessCount int
	ErrorCount   int
	Errors       []string
}

// ImportScheduleFromCSV импортирует расписание
// Формат CSV: GroupName;SubjectName;TeacherLastName;TeacherFirstName;TeacherPatronymic;Room;LessonType;StartsAt;EndsAt
func (s *ScheduleUseCase) ImportScheduleFromCSV(ctx context.Context, reader io.Reader, replaceExisting bool) (ImportResult, error) {
	result := ImportResult{}

	csvReader := csv.NewReader(reader)
	csvReader.Comma = ';'
	csvReader.TrimLeadingSpace = true

	// Пропускаем заголовок
	if _, err := csvReader.Read(); err != nil {
		return result, apperrors.ErrInvalidImportFormat
	}

	var lessons []models.Lesson
	groupDateRanges := make(map[uuid.UUID]*dateRange)

	lineNum := 1
	for {
		lineNum++
		record, err := csvReader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: %v", lineNum, err))
			result.ErrorCount++
			continue
		}

		result.TotalRows++

		lesson, err := s.parseScheduleRow(ctx, record, lineNum)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			result.ErrorCount++
			continue
		}

		lessons = append(lessons, lesson)

		if replaceExisting {
			updateDateRange(groupDateRanges, lesson.GroupID, lesson.StartsAt, lesson.EndsAt)
		}
	}

	if len(lessons) == 0 {
		return result, apperrors.ErrImportFailed
	}

	// Удаляем существующие занятия
	if replaceExisting {
		for groupID, dr := range groupDateRanges {
			deleted, _ := s.lessonRepo.DeleteByGroupAndDateRange(ctx, groupID, dr.from, dr.to)
			logger.Infof(ctx, "deleted %d lessons for group %s", deleted, groupID)
		}
	}

	// Массовая вставка
	inserted, err := s.lessonRepo.BulkCreate(ctx, lessons)
	if err != nil {
		logger.Errorf(ctx, "failed to bulk create: %v", err)
		return result, apperrors.ErrImportFailed
	}

	result.SuccessCount = inserted
	logger.Infof(ctx, "imported %d lessons", inserted)

	return result, nil
}

func (s *ScheduleUseCase) parseScheduleRow(ctx context.Context, record []string, lineNum int) (models.Lesson, error) {
	if len(record) < 9 {
		return models.Lesson{}, fmt.Errorf("line %d: expected 9 columns, got %d", lineNum, len(record))
	}

	groupName := strings.TrimSpace(record[0])
	subjectName := strings.TrimSpace(record[1])
	teacherLastName := strings.TrimSpace(record[2])
	teacherFirstName := strings.TrimSpace(record[3])
	teacherPatronymic := strings.TrimSpace(record[4])
	roomName := strings.TrimSpace(record[5])
	lessonType := strings.TrimSpace(record[6])
	startsAtStr := strings.TrimSpace(record[7])
	endsAtStr := strings.TrimSpace(record[8])

	// Группа должна существовать
	group, err := s.groupRepo.GetByName(ctx, groupName)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("line %d: group '%s' not found", lineNum, groupName)
	}

	// Предмет создаётся автоматически
	subjectID, err := s.subjectRepo.GetOrCreateByName(ctx, subjectName)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("line %d: subject error: %v", lineNum, err)
	}

	// Преподаватель создаётся автоматически
	teacherID, err := s.teacherRepo.GetOrCreateByFullName(ctx, teacherLastName, teacherFirstName, teacherPatronymic)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("line %d: teacher error: %v", lineNum, err)
	}

	// Аудитория создаётся автоматически
	roomID, err := s.classroomRepo.GetOrCreateByName(ctx, roomName)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("line %d: classroom error: %v", lineNum, err)
	}

	startsAt, err := time.Parse("2006-01-02 15:04", startsAtStr)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("line %d: invalid start time '%s'", lineNum, startsAtStr)
	}

	endsAt, err := time.Parse("2006-01-02 15:04", endsAtStr)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("line %d: invalid end time '%s'", lineNum, endsAtStr)
	}

	if !isValidLessonType(lessonType) {
		return models.Lesson{}, fmt.Errorf("line %d: invalid lesson type '%s'", lineNum, lessonType)
	}

	return models.Lesson{
		ID:         uuid.New(),
		GroupID:    group.ID,
		SubjectID:  subjectID,
		TeacherID:  teacherID,
		RoomID:     roomID,
		LessonType: models.LessonType(lessonType),
		StartsAt:   startsAt,
		EndsAt:     endsAt,
	}, nil
}

// ==================== Helpers ====================

func (s *ScheduleUseCase) validateLesson(ctx context.Context, lesson models.Lesson) error {
	if !lesson.EndsAt.After(lesson.StartsAt) {
		return apperrors.ErrInvalidLessonTime
	}

	if _, err := s.groupRepo.GetByID(ctx, lesson.GroupID); err != nil {
		return apperrors.ErrGroupNotFound
	}

	if _, err := s.subjectRepo.GetByID(ctx, lesson.SubjectID); err != nil {
		return apperrors.ErrSubjectNotFound
	}

	if _, err := s.teacherRepo.GetByID(ctx, lesson.TeacherID); err != nil {
		return apperrors.ErrTeacherNotFound
	}

	if _, err := s.classroomRepo.GetByID(ctx, lesson.RoomID); err != nil {
		return apperrors.ErrClassroomNotFound
	}

	return nil
}

func (s *ScheduleUseCase) enrichLessonDetails(ctx context.Context, lesson models.Lesson) models.LessonDetails {
	details := models.LessonDetails{Lesson: lesson}

	if group, err := s.groupRepo.GetByID(ctx, lesson.GroupID); err == nil {
		details.GroupName = group.Name
	}

	if teacher, err := s.teacherRepo.GetByID(ctx, lesson.TeacherID); err == nil {
		details.TeacherName = formatTeacherName(teacher)
	}

	if subject, err := s.subjectRepo.GetByID(ctx, lesson.SubjectID); err == nil {
		details.SubjectName = subject.Name
	}

	if classroom, err := s.classroomRepo.GetByID(ctx, lesson.RoomID); err == nil {
		details.RoomName = classroom.Name
	}

	if queue, err := s.queueRepo.GetByLessonID(ctx, lesson.ID); err == nil && queue != nil {
		details.QueueID = &queue.ID
	}

	return details
}

type dateRange struct {
	from, to time.Time
}

func updateDateRange(ranges map[uuid.UUID]*dateRange, groupID uuid.UUID, startsAt, endsAt time.Time) {
	if dr, ok := ranges[groupID]; ok {
		if startsAt.Before(dr.from) {
			dr.from = startsAt
		}
		if endsAt.After(dr.to) {
			dr.to = endsAt
		}
	} else {
		ranges[groupID] = &dateRange{from: startsAt, to: endsAt}
	}
}

func getWeekStart(date time.Time) time.Time {
	weekday := date.Weekday()
	if weekday == time.Sunday {
		date = date.AddDate(0, 0, -6)
	} else {
		date = date.AddDate(0, 0, -int(weekday)+1)
	}
	return time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
}

func formatTeacherName(t models.Teacher) string {
	if t.Patronymic != "" {
		return t.LastName + " " + string([]rune(t.FirstName)[0]) + "." + string([]rune(t.Patronymic)[0]) + "."
	}
	return t.LastName + " " + string([]rune(t.FirstName)[0]) + "."
}

func isValidLessonType(t string) bool {
	switch models.LessonType(t) {
	case models.LessonTypeLecture, models.LessonTypeSeminar, models.LessonTypeLab:
		return true
	}
	return false
}
