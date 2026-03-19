package gormrepository

import (
	"context"

	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	gormmodels "stud_hub/internal/repository/gorm-models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type QueueRepository struct {
	db        *gorm.DB
}

func NewQueueRepository(db *gorm.DB) *QueueRepository {
	return &QueueRepository{db: db}
}

// Queue CRUD

func (r *QueueRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Queue, error) {
	var q gormmodels.Queue
	if err := r.db.WithContext(ctx).First(&q, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.Queue{}, apperrors.ErrQueueNotFound
		}
		logger.Errorf(ctx, "gorm: failed to get queue %s: %v", id, err)
		return models.Queue{}, apperrors.ErrInternalServer
	}
	return gormmodels.FromGormQueue(q), nil
}

// func (r *QueueRepository) GetByIDWithSlots(ctx context.Context, id uuid.UUID) (models.Queue, error) {
// 	queue, err := r.GetByID(ctx, id)
// 	if err != nil {
// 		return models.Queue{}, err
// 	}

// 	slots, err := r.slotsRepo.GetSlotsByQueueID(ctx, id)
// 	if err != nil {
// 		return models.Queue{}, err
// 	}
// 	queue.Slots = toSlotPointers(slots)
// 	return queue, nil
// }

func (r *QueueRepository) GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.Queue, error) {
	var queues []gormmodels.Queue
	if err := r.db.WithContext(ctx).Where("group_id = ?", groupID).
		Order("created_at desc").
		Find(&queues).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to get queues for group %s: %v", groupID, err)
		return nil, apperrors.ErrInternalServer
	}

	result := make([]models.Queue, 0, len(queues))
	for _, q := range queues {
		result = append(result, gormmodels.FromGormQueue(q))
	}
	return result, nil
}

func (r *QueueRepository) GetByLessonID(ctx context.Context, lessonID uuid.UUID) (*models.Queue, error) {
	var q gormmodels.Queue
	err := r.db.WithContext(ctx).Where("lesson_id = ?", lessonID).First(&q).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrQueueNotFound
	}
	if err != nil {
		logger.Errorf(ctx, "gorm: failed to get queue by lesson %s: %v", lessonID, err)
		return nil, apperrors.ErrInternalServer
	}
	queue := gormmodels.FromGormQueue(q)
	return &queue, nil
}

func (r *QueueRepository) GetActiveByGroupAndSubject(ctx context.Context, groupID, subjectID uuid.UUID) (*models.Queue, error) {
	var q gormmodels.Queue
	err := r.db.WithContext(ctx).
		Where("group_id = ? AND subject_id = ? AND status = ?", groupID, subjectID, gormmodels.QueueStatusOpen).
		Order("opens_at desc").
		First(&q).Error
	if err == gorm.ErrRecordNotFound {
		return nil, apperrors.ErrQueueNotFound
	}
	if err != nil {
		logger.Errorf(ctx, "gorm: failed to get active queue for group %s subject %s: %v", groupID, subjectID, err)
		return nil, apperrors.ErrInternalServer
	}
	queue := gormmodels.FromGormQueue(q)
	return &queue, nil
}

func (r *QueueRepository) Create(ctx context.Context, queue models.Queue) (uuid.UUID, error) {
	gQueue := gormmodels.ToGormQueue(queue)
	if err := r.db.WithContext(ctx).Create(&gQueue).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create queue %s: %v", queue.ID, err)
		return uuid.UUID{}, apperrors.ErrInternalServer
	}
	return gQueue.ID, nil
}

func (r *QueueRepository) Update(ctx context.Context, queue models.Queue) error {
	gQueue := gormmodels.ToGormQueue(queue)

	builder := PartialUpdateBuilder().
		UpdateUUID("group_id", gQueue.GroupID).
		UpdateUUID("subject_id", gQueue.SubjectID).
		SetPtrUUID("lesson_id", gQueue.LessonID).
		UpdateUUID("created_by", gQueue.CreatedBy).
		UpdateTime("created_at", gQueue.CreatedAt).
		UpdateTime("opens_at", gQueue.OpensAt).
		UpdateValue("closes_at", gQueue.ClosesAt).
		UpdateValue("max_size", gQueue.MaxSize).
		UpdateValue("status", gQueue.Status)

	updates := builder.Build()
	if len(updates) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).Model(&gormmodels.Queue{ID: queue.ID}).Updates(updates).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to update queue %s: %v", queue.ID, err)
		return apperrors.ErrInternalServer
	}
	return nil
}

func (r *QueueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).Delete(&gormmodels.Queue{}, "id = ?", id).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to delete queue %s: %v", id, err)
		return apperrors.ErrInternalServer
	}
	return nil
}

