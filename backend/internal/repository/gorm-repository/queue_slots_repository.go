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

type QueueSlotsRepository struct {
	db *gorm.DB
}

func NewQueueSlotsRepository(db *gorm.DB) *QueueSlotsRepository {
	return &QueueSlotsRepository{db: db}
}

func (r *QueueSlotsRepository) GetSlotByID(ctx context.Context, id uuid.UUID) (models.QueueSlot, error) {
	var slot gormmodels.QueueSlot
	if err := r.db.WithContext(ctx).First(&slot, "id = ?", id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return models.QueueSlot{}, apperrors.ErrSlotNotFound
		}
		logger.Errorf(ctx, "gorm: failed to get slot %s: %v", id, err)
		return models.QueueSlot{}, apperrors.ErrInternalServer
	}
	result, err := r.attachPosition(ctx, slot)
	if err != nil {
		return models.QueueSlot{}, err
	}
	return result, nil
}

func (r *QueueSlotsRepository) GetSlotByQueueAndStudent(ctx context.Context, queueID, studentID uuid.UUID) (*models.QueueSlot, error) {
	var slot gormmodels.QueueSlot
	err := r.db.WithContext(ctx).
		Where("queue_id = ? AND student_id = ?", queueID, studentID).
		First(&slot).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	if err != nil {
		logger.Errorf(ctx, "gorm: failed to get slot by queue %s and student %s: %v", queueID, studentID, err)
		return nil, apperrors.ErrInternalServer
	}
	res, err := r.attachPosition(ctx, slot)
	if err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *QueueSlotsRepository) GetSlotsByQueueID(ctx context.Context, queueID uuid.UUID) ([]models.QueueSlot, error) {
	var slots []gormmodels.QueueSlot
	if err := r.db.WithContext(ctx).
		Where("queue_id = ?", queueID).
		Order("signed_up_at asc, id asc").
		Find(&slots).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to list slots for queue %s: %v", queueID, err)
		return nil, apperrors.ErrInternalServer
	}

	return r.attachPositions(slots), nil
}

func (r *QueueSlotsRepository) GetSlotsCount(ctx context.Context, queueID uuid.UUID) (int, error) {
	var count int64
	if err := r.db.WithContext(ctx).
		Model(&gormmodels.QueueSlot{}).
		Where("queue_id = ?", queueID).
		Count(&count).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to count slots for queue %s: %v", queueID, err)
		return 0, apperrors.ErrInternalServer
	}
	return int(count), nil
}

func (r *QueueSlotsRepository) CreateSlot(ctx context.Context, slot models.QueueSlot) (*models.QueueSlot, error) {
	if slot.SignedUpAt.IsZero() {
		slot.SignedUpAt = time.Now()
	}

	gSlot := toGormSlot(slot)
	if err := r.db.WithContext(ctx).Create(&gSlot).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to create slot: %v", err)
		return nil, apperrors.ErrInternalServer
	}

	result, err := r.attachPosition(ctx, gSlot)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (r *QueueSlotsRepository) CreateSlots(ctx context.Context, slots []models.QueueSlot) (int, error) {
	if len(slots) == 0 {
		return 0, nil
	}

	now := time.Now()
	gSlots := make([]gormmodels.QueueSlot, 0, len(slots))
	for _, s := range slots {
		if s.SignedUpAt.IsZero() {
			s.SignedUpAt = now
		}
		gSlots = append(gSlots, toGormSlot(s))
	}

	if err := r.db.WithContext(ctx).Create(&gSlots).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to bulk create slots: %v", err)
		return 0, apperrors.ErrInternalServer
	}
	return len(gSlots), nil
}

func (r *QueueSlotsRepository) UpdateSlot(ctx context.Context, slot models.QueueSlot) error {
	gSlot := toGormSlot(slot)

	builder := PartialUpdateBuilder().
		UpdateValue("status", gSlot.Status).
		UpdateTime("signed_up_at", gSlot.SignedUpAt)

	updates := builder.Build()
	if len(updates) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).
		Model(&gormmodels.QueueSlot{ID: slot.ID}).
		Updates(updates).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to update slot %s: %v", slot.ID, err)
		return apperrors.ErrInternalServer
	}
	return nil
}

func (r *QueueSlotsRepository) DeleteSlot(ctx context.Context, id uuid.UUID) error {
	if err := r.db.WithContext(ctx).
		Delete(&gormmodels.QueueSlot{}, "id = ?", id).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to delete slot %s: %v", id, err)
		return apperrors.ErrInternalServer
	}
	return nil
}

func (r *QueueSlotsRepository) GetFailedSlotsByQueueID(ctx context.Context, queueID uuid.UUID) ([]models.QueueSlot, error) {
	var slots []gormmodels.QueueSlot
	if err := r.db.WithContext(ctx).
		Where("queue_id = ? AND status IN (?, ?)", queueID, gormmodels.QueueSlotStatusFailed, gormmodels.QueueSlotStatusNoShow).
		Order("signed_up_at asc, id asc").
		Find(&slots).Error; err != nil {
		logger.Errorf(ctx, "gorm: failed to get failed slots for queue %s: %v", queueID, err)
		return nil, apperrors.ErrInternalServer
	}
	return r.attachPositions(slots), nil
}

func (r *QueueSlotsRepository) GetLastPosition(ctx context.Context, queueID uuid.UUID) (int, error) {
	count, err := r.GetSlotsCount(ctx, queueID)
	if err != nil {
		return 0, err
	}
	return count, nil
}

// helpers
func (r *QueueSlotsRepository) attachPosition(ctx context.Context, slot gormmodels.QueueSlot) (models.QueueSlot, error) {
	pos, err := r.computePosition(ctx, slot)
	if err != nil {
		return models.QueueSlot{}, err
	}
	result := fromGormSlot(slot)
	result.Position = pos
	return result, nil
}

func (r *QueueSlotsRepository) attachPositions(slots []gormmodels.QueueSlot) []models.QueueSlot {
	result := make([]models.QueueSlot, 0, len(slots))
	for i, s := range slots {
		slot := fromGormSlot(s)
		slot.Position = i + 1 // ordered by signed_up_at,id
		result = append(result, slot)
	}
	return result
}

func (r *QueueSlotsRepository) computePosition(ctx context.Context, slot gormmodels.QueueSlot) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&gormmodels.QueueSlot{}).
		Where("queue_id = ? AND (signed_up_at < ? OR (signed_up_at = ? AND id <= ?))",
			slot.QueueID, slot.SignedUpAt, slot.SignedUpAt, slot.ID).
		Count(&count).Error
	if err != nil {
		logger.Errorf(ctx, "gorm: failed to compute position for slot %s: %v", slot.ID, err)
		return 0, apperrors.ErrInternalServer
	}
	return int(count), nil
}

// converters
func toGormSlot(s models.QueueSlot) gormmodels.QueueSlot {
	return gormmodels.QueueSlot{
		ID:         s.ID,
		QueueID:    s.QueueID,
		StudentID:  s.StudentID,
		Status:     gormmodels.QueueSlotStatus(s.Status),
		SignedUpAt: s.SignedUpAt,
	}
}

func fromGormSlot(s gormmodels.QueueSlot) models.QueueSlot {
	return models.QueueSlot{
		ID:         s.ID,
		QueueID:    s.QueueID,
		StudentID:  s.StudentID,
		Status:     models.SlotStatus(s.Status),
		SignedUpAt: s.SignedUpAt,
	}
}
