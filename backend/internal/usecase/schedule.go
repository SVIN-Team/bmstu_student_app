package usecase

import (
	"context"
	"fmt"
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
	GetLessonDetails(ctx context.Context, id uuid.UUID) (models.LessonDetails, error)
}

type SubjectRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Subject, error)
	GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error)
}

type TeacherRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Teacher, error)
	GetOrCreateByFullName(ctx context.Context, lastName, firstName, patronymic string) (uuid.UUID, error)
}

type ClassroomRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Classroom, error)
	GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error)
}

type ScheduleUseCase struct {
	lessonRepo    LessonRepository
	subjectRepo   SubjectRepository
	groupRepo     GroupRepository
	teacherRepo   TeacherRepository
	classroomRepo ClassroomRepository
	queueRepo     QueueRepository
}

func NewScheduleUseCase(
	lessonRepo LessonRepository,
	subjectRepo SubjectRepository,
	groupRepo GroupRepository,
	teacherRepo TeacherRepository,
	classroomRepo ClassroomRepository,
	queueRepo QueueRepository,
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
		details, err := s.enrichLessonDetails(ctx, lesson)
		if err != nil {
			logger.Errorf(ctx, "failed to enrich lesson details: %v", err)
			return nil, apperrors.ErrInternalServer
		}
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
	return s.enrichLessonDetails(ctx, lesson)
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

func (s *ScheduleUseCase) ImportSchedule(ctx context.Context, rows []models.ScheduleImportRow, replaceExisting bool) (ImportResult, error) {
	result := ImportResult{TotalRows: len(rows)}
	var lessons []models.Lesson
	groupDateRanges := make(map[uuid.UUID]*dateRange)

	for _, row := range rows {
		// Преобразуем строку-контракт в доменную модель Lesson
		lesson, err := s.processImportRow(ctx, row)
		if err != nil {
			result.Errors = append(result.Errors, err.Error())
			result.ErrorCount++
			continue
		}

		lessons = append(lessons, lesson)

		// Собираем диапазоны дат для удаления, если нужно
		if replaceExisting {
			updateDateRange(groupDateRanges, lesson.GroupID, lesson.StartsAt, lesson.EndsAt)
		}
	}

	if result.ErrorCount > 0 {
		return result, apperrors.ErrImportFailed
	}

	if len(lessons) == 0 {
		// Если после валидации не осталось ни одного занятия для вставки
		return result, nil
	}

	// Удаляем существующие занятия в нужных диапазонах
	if replaceExisting {
		for groupID, dr := range groupDateRanges {
			deleted, err := s.lessonRepo.DeleteByGroupAndDateRange(ctx, groupID, dr.from, dr.to)
			if err != nil {
				logger.Errorf(ctx, "failed to delete existing lessons for group %s in range %v - %v: %v", groupID, dr.from, dr.to, err)
				return result, apperrors.ErrInternalServer
			}
			logger.Infof(ctx, "deleted %d existing lessons for group %s", deleted, groupID)
		}
	}

	// Массовая вставка
	inserted, err := s.lessonRepo.BulkCreate(ctx, lessons)
	if err != nil {
		logger.Errorf(ctx, "failed to bulk create lessons: %v", err)
		return result, apperrors.ErrInternalServer
	}

	result.SuccessCount = inserted
	logger.Infof(ctx, "successfully imported %d lessons", inserted)

	return result, nil
}

func (s *ScheduleUseCase) processImportRow(ctx context.Context, row models.ScheduleImportRow) (models.Lesson, error) {
	if !row.EndsAt.After(row.StartsAt) {
		return models.Lesson{}, fmt.Errorf("line %d: end time must be after start time", row.LineNum)
	}

	// Группа создаётся или находится автоматически
	groupID, err := s.groupRepo.GetOrCreateByName(ctx, row.GroupName)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("line %d: group error: %v", row.LineNum, err)
	}

	// Предмет создаётся или находится автоматически
	subjectID, err := s.subjectRepo.GetOrCreateByName(ctx, row.SubjectName)
	if err != nil {
		return models.Lesson{}, fmt.Errorf("line %d: subject error: %v", row.LineNum, err)
	}

	var teacherID *uuid.UUID
	if row.TeacherLastName != "" && row.TeacherFirstName != "" && row.TeacherPatronymic != "" {
		id, err := s.teacherRepo.GetOrCreateByFullName(ctx, row.TeacherLastName, row.TeacherFirstName, row.TeacherPatronymic)
		if err != nil {
			return models.Lesson{}, fmt.Errorf("line %d: teacher error: %v", row.LineNum, err)
		}
		teacherID = &id
	}

	var roomID *uuid.UUID
	if row.RoomName != "" {
		id, err := s.classroomRepo.GetOrCreateByName(ctx, row.RoomName)
		if err != nil {
			return models.Lesson{}, fmt.Errorf("line %d: classroom error: %v", row.LineNum, err)
		}
		roomID = &id
	}

	var teacherIDVal uuid.UUID
	if teacherID != nil {
		teacherIDVal = *teacherID
	} else {
		teacherIDVal = uuid.Nil
	}

	var roomIDVal uuid.UUID
	if roomID != nil {
		roomIDVal = *roomID
	} else {
		roomIDVal = uuid.Nil
	}

	return models.Lesson{
		ID:         uuid.New(),
		GroupID:    groupID,
		SubjectID:  subjectID,
		TeacherID:  teacherIDVal,
		RoomID:     roomIDVal,
		LessonType: row.LessonType,
		StartsAt:   row.StartsAt,
		EndsAt:     row.EndsAt,
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

	if lesson.TeacherID != uuid.Nil {
		if _, err := s.teacherRepo.GetByID(ctx, lesson.TeacherID); err != nil {
			return apperrors.ErrTeacherNotFound
		}
	}

	if lesson.RoomID != uuid.Nil {
		if _, err := s.classroomRepo.GetByID(ctx, lesson.RoomID); err != nil {
			return apperrors.ErrClassroomNotFound
		}
	}

	return nil
}

func (s *ScheduleUseCase) enrichLessonDetails(ctx context.Context, lesson models.Lesson) (models.LessonDetails, error) {
	details, err := s.lessonRepo.GetLessonDetails(ctx, lesson.ID)
	if err != nil {
		logger.Errorf(ctx, "failed to get lesson details: %v", err)
		return details, err
	}

	return details, nil
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
