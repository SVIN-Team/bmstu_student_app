package gormrepository

import (
	"context"
	"time"

	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	gormmodels "stud_hub/internal/repository/gorm-models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type LessonRepository struct {
	db *gorm.DB
}

func NewLessonRepository(db *gorm.DB) *LessonRepository {
	return &LessonRepository{db: db}
}

func (r *LessonRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Lesson, error) {
	var l gormmodels.Lesson
	if err := r.db.WithContext(ctx).First(&l, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Lesson{}, apperrors.ErrLessonNotFound
		}
		logger.Errorf(ctx, "gorm: failed to get lesson %s: %v", id, err)
		return models.Lesson{}, apperrors.ErrInternalServer
	}
	return fromGormLesson(l), nil
}

func (r *LessonRepository) GetByGroupAndDateRange(ctx context.Context, groupID uuid.UUID, from, to time.Time) ([]models.Lesson, error) {
	var lessons []gormmodels.Lesson
	if err := r.db.WithContext(ctx).
		Where("group_id = ? AND starts_at >= ? AND starts_at < ?", groupID, from, to).
		Order("starts_at asc").
		Find(&lessons).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to get lessons for group %s: %v", groupID, err)
		return nil, apperrors.ErrInternalServer
	}

	result := make([]models.Lesson, 0, len(lessons))
	for _, l := range lessons {
		result = append(result, fromGormLesson(l))
	}
	return result, nil
}

func (r *LessonRepository) Create(ctx context.Context, lesson models.Lesson) (uuid.UUID, error) {
	l := toGormLesson(lesson)
	if err := r.db.WithContext(ctx).Create(&l).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create lesson %s: %v", lesson.ID, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}
	return l.ID, nil
}

func (r *LessonRepository) Update(ctx context.Context, lesson models.Lesson) error {
	l := toGormLesson(lesson)
	builder := PartialUpdateBuilder()
	builder.
		UpdateUUID("group_id", l.GroupID).
		UpdateUUID("subject_id", l.SubjectID).
		UpdateUUID("teacher_id", l.TeacherID).
		UpdatePtrUUID("room_id", l.RoomID).
		UpdateValue("type", l.Type).
		UpdateTime("starts_at", l.StartsAt).
		UpdateTime("ends_at", l.EndsAt)

	updates := builder.Build()

	if len(updates) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).Model(&gormmodels.Lesson{ID: lesson.ID}).Updates(updates).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to update lesson %s: %v", lesson.ID, err)
		return apperrors.ErrInternalServer
	}
	return nil
}

func (r *LessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&gormmodels.Lesson{}, "id = ?", id).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to delete lesson %s: %v", id, err)
		return apperrors.ErrInternalServer
	}
	return nil
}

func (r *LessonRepository) CheckOverlap(ctx context.Context, groupID uuid.UUID, startsAt, endsAt time.Time, excludeID *uuid.UUID) (bool, error) {
	query := r.db.WithContext(ctx).Model(&gormmodels.Lesson{}).
		Where("group_id = ? AND starts_at < ? AND ends_at > ?", groupID, endsAt, startsAt)

	if excludeID != nil {
		query = query.Where("id <> ?", *excludeID)
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to check overlap for group %s: %v", groupID, err)
		return false, apperrors.ErrInternalServer
	}
	return count > 0, nil
}

func (r *LessonRepository) BulkCreate(ctx context.Context, lessons []models.Lesson) (int, error) {
	gLessons := make([]gormmodels.Lesson, 0, len(lessons))
	for _, l := range lessons {
		gLessons = append(gLessons, toGormLesson(l))
	}

	if err := r.db.WithContext(ctx).CreateInBatches(&gLessons, 100).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to bulk create lessons: %v", err)
		return 0, apperrors.ErrInternalServer
	}
	return len(gLessons), nil
}

func (r *LessonRepository) DeleteByGroupAndDateRange(ctx context.Context, groupID uuid.UUID, from, to time.Time) (int, error) {
	tx := r.db.WithContext(ctx).Where("group_id = ? AND starts_at >= ? AND starts_at < ?", groupID, from, to).
		Delete(&gormmodels.Lesson{})
	if err := tx.Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to delete lessons for group %s: %v", groupID, err)
		return 0, apperrors.ErrInternalServer
	}
	return int(tx.RowsAffected), nil
}

func (r *LessonRepository) GetLessonDetails(ctx context.Context, id uuid.UUID) (models.LessonDetails, error) {
	var l gormmodels.Lesson
	if err := r.db.WithContext(ctx).
		Preload("Group").
		Preload("Subject").
		Preload("Teacher").
		Preload("Room").
		First(&l, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.LessonDetails{}, apperrors.ErrLessonNotFound
		}
		logger.Errorf(ctx, "gorm: failed to load lesson details %s: %v", id, err)
		return models.LessonDetails{}, apperrors.ErrInternalServer
	}

	lesson := fromGormLesson(l)

	roomName := ""
	if l.Room != nil && l.Room.Name != nil {
		roomName = *l.Room.Name
	}

	queueID, err := r.getQueueIDByLesson(ctx, lesson.ID)
	if err != nil {
		return models.LessonDetails{}, err
	}

	return models.LessonDetails{
		Lesson:      lesson,
		GroupName:   l.Group.Name,
		TeacherName: l.Teacher.Name(),
		SubjectName: l.Subject.Name,
		RoomName:    roomName,
		QueueID:     queueID,
	}, nil
}

func (r *LessonRepository) getQueueIDByLesson(ctx context.Context, lessonID uuid.UUID) (*uuid.UUID, error) {
	var q gormmodels.Queue
	err := r.db.WithContext(ctx).Select("id").Where("lesson_id = ?", lessonID).Order("created_at desc").First(&q).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		logger.Errorf(ctx, "gorm: failed to get queue for lesson %s: %v", lessonID, err)
		return nil, apperrors.ErrInternalServer
	}
	return &q.ID, nil
}

// converters
func toGormLesson(l models.Lesson) gormmodels.Lesson {
	var roomID *uuid.UUID
	if l.RoomID != uuid.Nil {
		roomID = &l.RoomID
	}
	return gormmodels.Lesson{
		ID:        l.ID,
		GroupID:   l.GroupID,
		SubjectID: l.SubjectID,
		TeacherID: l.TeacherID,
		RoomID:    roomID,
		Type:      gormmodels.LessonType(l.LessonType),
		StartsAt:  l.StartsAt,
		EndsAt:    l.EndsAt,
	}
}

func fromGormLesson(l gormmodels.Lesson) models.Lesson {
	var roomID uuid.UUID
	if l.RoomID != nil {
		roomID = *l.RoomID
	}

	return models.Lesson{
		ID:         l.ID,
		GroupID:    l.GroupID,
		TeacherID:  l.TeacherID,
		SubjectID:  l.SubjectID,
		RoomID:     roomID,
		LessonType: models.LessonType(l.Type),
		StartsAt:   l.StartsAt,
		EndsAt:     l.EndsAt,
	}
}
