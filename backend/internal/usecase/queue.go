// internal/usecase/queue_usecase.go
package usecase

import (
	"context"
	"time"

	apperrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
)

type QueueRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.Queue, error)
	GetByIDWithSlots(ctx context.Context, id uuid.UUID) (models.Queue, error)
	GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.Queue, error)
	GetByLessonID(ctx context.Context, lessonID uuid.UUID) (*models.Queue, error)
	GetActiveByGroupAndSubject(ctx context.Context, groupID, subjectID uuid.UUID) (*models.Queue, error)
	Create(ctx context.Context, queue models.Queue) (uuid.UUID, error)
	Update(ctx context.Context, queue models.Queue) error
	Delete(ctx context.Context, id uuid.UUID) error

	// Slots
	GetSlotByID(ctx context.Context, id uuid.UUID) (models.QueueSlot, error)
	GetSlotByQueueAndStudent(ctx context.Context, queueID, studentID uuid.UUID) (*models.QueueSlot, error)
	GetSlotsByQueueID(ctx context.Context, queueID uuid.UUID) ([]models.QueueSlot, error)
	GetSlotsCount(ctx context.Context, queueID uuid.UUID) (int, error)
	CreateSlot(ctx context.Context, slot models.QueueSlot) (uuid.UUID, error)
	CreateSlots(ctx context.Context, slots []models.QueueSlot) (int, error)
	UpdateSlot(ctx context.Context, slot models.QueueSlot) error
	DeleteSlot(ctx context.Context, id uuid.UUID) error
	GetFailedSlotsByQueueID(ctx context.Context, queueID uuid.UUID) ([]models.QueueSlot, error)
	GetLastPosition(ctx context.Context, queueID uuid.UUID) (int, error)
}

type UserRepositoryForQueue interface {
	GetByID(ctx context.Context, id uuid.UUID) (models.User, error)
}

type QueueUseCase struct {
	queueRepo QueueRepository
	userRepo  UserRepositoryForQueue
}

func NewQueueUseCase(queueRepo QueueRepository, userRepo UserRepositoryForQueue) *QueueUseCase {
	return &QueueUseCase{
		queueRepo: queueRepo,
		userRepo:  userRepo,
	}
}

// ==================== Чтение (для всех авторизованных) ====================

// GetByID возвращает очередь по ID
func (q *QueueUseCase) GetByID(ctx context.Context, queueID uuid.UUID) (models.Queue, error) {
	queue, err := q.queueRepo.GetByIDWithSlots(ctx, queueID)
	if err != nil {
		return models.Queue{}, apperrors.ErrQueueNotFound
	}
	return queue, nil
}

// GetByGroupID возвращает все очереди группы
func (q *QueueUseCase) GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.Queue, error) {
	queues, err := q.queueRepo.GetByGroupID(ctx, groupID)
	if err != nil {
		logger.Errorf(ctx, "failed to get queues for group %s: %v", groupID, err)
		return nil, apperrors.ErrInternalServer
	}
	return queues, nil
}

// GetActiveByGroupAndSubject возвращает активную очередь по группе и предмету
func (q *QueueUseCase) GetActiveByGroupAndSubject(ctx context.Context, groupID, subjectID uuid.UUID) (*models.Queue, error) {
	queue, err := q.queueRepo.GetActiveByGroupAndSubject(ctx, groupID, subjectID)
	if err != nil {
		return nil, apperrors.ErrQueueNotFound
	}
	return queue, nil
}

// GetMyQueues возвращает очереди, в которых записан студент
func (q *QueueUseCase) GetMyQueues(ctx context.Context, studentID uuid.UUID) ([]models.Queue, error) {
	user, err := q.userRepo.GetByID(ctx, studentID)
	if err != nil {
		return nil, apperrors.ErrUserNotFound
	}

	if user.GroupID == uuid.Nil {
		return nil, apperrors.ErrNotGroupMember
	}

	queues, err := q.queueRepo.GetByGroupID(ctx, user.GroupID)
	if err != nil {
		logger.Errorf(ctx, "failed to get queues: %v", err)
		return nil, apperrors.ErrInternalServer
	}

	// Фильтруем только те, где студент записан
	var result []models.Queue
	for _, queue := range queues {
		slot, _ := q.queueRepo.GetSlotByQueueAndStudent(ctx, queue.ID, studentID)
		if slot != nil {
			result = append(result, queue)
		}
	}

	return result, nil
}

// ==================== Создание/управление (для старосты) ====================

// CreateQueueParams параметры создания очереди
type CreateQueueParams struct {
	GroupID            uuid.UUID
	SubjectID          uuid.UUID
	LessonID           uuid.UUID
	CreatedByUserID    uuid.UUID
	OpensAt            time.Time
	ClosesAt           *time.Time
	MaxSize            *uint32
	TransferFailedFrom *uuid.UUID // ID предыдущей очереди для переноса неуспевших
}

// Create создаёт новую очередь
func (q *QueueUseCase) Create(ctx context.Context, params CreateQueueParams) (uuid.UUID, error) {
	// Валидация времени
	if params.ClosesAt != nil && !params.ClosesAt.After(params.OpensAt) {
		return uuid.Nil, apperrors.ErrInvalidQueueTime
	}

	queue := models.Queue{
		ID:              uuid.New(),
		GroupID:         params.GroupID,
		SubjectID:       params.SubjectID,
		LessonID:        params.LessonID,
		CreatedByUserID: params.CreatedByUserID,
		CreatedAt:       time.Now(),
		OpensAt:         params.OpensAt,
		ClosesAt:        params.ClosesAt,
		MaxSize:         params.MaxSize,
		Status:          models.QueueStatusDraft,
	}

	id, err := q.queueRepo.Create(ctx, queue)
	if err != nil {
		logger.Errorf(ctx, "failed to create queue: %v", err)
		return uuid.Nil, apperrors.ErrInternalServer
	}

	// Переносим неуспевших из предыдущей очереди
	if params.TransferFailedFrom != nil {
		transferred, err := q.transferFailedStudents(ctx, *params.TransferFailedFrom, id)
		if err != nil {
			logger.Warnf(ctx, "failed to transfer students: %v", err)
		} else {
			logger.Infof(ctx, "transferred %d failed students to queue %s", transferred, id)
		}
	}

	logger.Infof(ctx, "queue created: %s by user %s", id, params.CreatedByUserID)
	return id, nil
}

// Update обновляет параметры очереди
func (q *QueueUseCase) Update(ctx context.Context, headmanID uuid.UUID, queue models.Queue) error {
	existing, err := q.queueRepo.GetByID(ctx, queue.ID)
	if err != nil {
		return apperrors.ErrQueueNotFound
	}

	// Проверяем, что это создатель очереди
	if existing.CreatedByUserID != headmanID {
		return apperrors.ErrForbidden
	}

	// Нельзя редактировать закрытую очередь
	if existing.Status == models.QueueStatusClosed {
		return apperrors.ErrQueueClosed
	}

	// Валидация времени
	if queue.ClosesAt != nil && !queue.ClosesAt.After(queue.OpensAt) {
		return apperrors.ErrInvalidQueueTime
	}

	// Сохраняем неизменяемые поля
	queue.GroupID = existing.GroupID
	queue.SubjectID = existing.SubjectID
	queue.LessonID = existing.LessonID
	queue.CreatedByUserID = existing.CreatedByUserID
	queue.CreatedAt = existing.CreatedAt

	if err := q.queueRepo.Update(ctx, queue); err != nil {
		logger.Errorf(ctx, "failed to update queue: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "queue updated: %s", queue.ID)
	return nil
}

// Delete удаляет очередь
func (q *QueueUseCase) Delete(ctx context.Context, headmanID, queueID uuid.UUID) error {
	queue, err := q.queueRepo.GetByID(ctx, queueID)
	if err != nil {
		return apperrors.ErrQueueNotFound
	}

	if queue.CreatedByUserID != headmanID {
		return apperrors.ErrForbidden
	}

	if err := q.queueRepo.Delete(ctx, queueID); err != nil {
		logger.Errorf(ctx, "failed to delete queue: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "queue deleted: %s", queueID)
	return nil
}

// Open открывает очередь для записи
func (q *QueueUseCase) Open(ctx context.Context, headmanID, queueID uuid.UUID) error {
	queue, err := q.queueRepo.GetByID(ctx, queueID)
	if err != nil {
		return apperrors.ErrQueueNotFound
	}

	if queue.CreatedByUserID != headmanID {
		return apperrors.ErrForbidden
	}

	if queue.Status != models.QueueStatusDraft {
		return apperrors.ErrQueueAlreadyOpen
	}

	queue.Status = models.QueueStatusOpen

	if err := q.queueRepo.Update(ctx, queue); err != nil {
		logger.Errorf(ctx, "failed to open queue: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "queue opened: %s", queueID)
	return nil
}

// Close закрывает очередь
func (q *QueueUseCase) Close(ctx context.Context, headmanID, queueID uuid.UUID) error {
	queue, err := q.queueRepo.GetByID(ctx, queueID)
	if err != nil {
		return apperrors.ErrQueueNotFound
	}

	if queue.CreatedByUserID != headmanID {
		return apperrors.ErrForbidden
	}

	if queue.Status == models.QueueStatusClosed {
		return apperrors.ErrQueueClosed
	}

	queue.Status = models.QueueStatusClosed
	now := time.Now()
	queue.ClosesAt = &now

	if err := q.queueRepo.Update(ctx, queue); err != nil {
		logger.Errorf(ctx, "failed to close queue: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "queue closed: %s", queueID)
	return nil
}

// ==================== Запись в очередь (для студентов) ====================

// SignUp записывает студента в очередь
func (q *QueueUseCase) SignUp(ctx context.Context, studentID, queueID uuid.UUID) (int, error) {
	queue, err := q.queueRepo.GetByID(ctx, queueID)
	if err != nil {
		return 0, apperrors.ErrQueueNotFound
	}

	// Проверяем статус очереди
	if queue.Status != models.QueueStatusOpen {
		return 0, apperrors.ErrQueueNotOpen
	}

	// Проверяем время
	now := time.Now()
	if now.Before(queue.OpensAt) {
		return 0, apperrors.ErrQueueNotOpen
	}
	if queue.ClosesAt != nil && now.After(*queue.ClosesAt) {
		return 0, apperrors.ErrQueueClosed
	}

	// Проверяем, что студент из той же группы
	user, err := q.userRepo.GetByID(ctx, studentID)
	if err != nil {
		return 0, apperrors.ErrUserNotFound
	}
	if user.GroupID != queue.GroupID {
		return 0, apperrors.ErrNotInQueueGroup
	}

	// Проверяем, не записан ли уже
	existingSlot, _ := q.queueRepo.GetSlotByQueueAndStudent(ctx, queueID, studentID)
	if existingSlot != nil {
		return 0, apperrors.ErrAlreadyInQueue
	}

	// Проверяем лимит
	if queue.MaxSize != nil {
		count, err := q.queueRepo.GetSlotsCount(ctx, queueID)
		if err != nil {
			logger.Errorf(ctx, "failed to get slots count: %v", err)
			return 0, apperrors.ErrInternalServer
		}
		if count >= int(*queue.MaxSize) {
			return 0, apperrors.ErrQueueFull
		}
	}

	// Получаем следующую позицию
	lastPosition, err := q.queueRepo.GetLastPosition(ctx, queueID)
	if err != nil {
		logger.Errorf(ctx, "failed to get last position: %v", err)
		return 0, apperrors.ErrInternalServer
	}
	position := lastPosition + 1

	slot := models.QueueSlot{
		ID:         uuid.New(),
		QueueID:    queueID,
		StudentID:  studentID,
		Status:     models.SlotStatusWaiting,
		SignedUpAt: time.Now(),
	}

	if _, err := q.queueRepo.CreateSlot(ctx, slot); err != nil {
		logger.Errorf(ctx, "failed to create slot: %v", err)
		return 0, apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "student %s signed up for queue %s at position %d", studentID, queueID, position)
	return position, nil
}

// CancelSignUp отменяет запись студента
func (q *QueueUseCase) CancelSignUp(ctx context.Context, studentID, queueID uuid.UUID) error {
	queue, err := q.queueRepo.GetByID(ctx, queueID)
	if err != nil {
		return apperrors.ErrQueueNotFound
	}

	// Нельзя отменить запись в закрытой очереди
	if queue.Status == models.QueueStatusClosed {
		return apperrors.ErrQueueClosed
	}

	slot, err := q.queueRepo.GetSlotByQueueAndStudent(ctx, queueID, studentID)
	if err != nil || slot == nil {
		return apperrors.ErrNotInQueue
	}

	if err := q.queueRepo.DeleteSlot(ctx, slot.ID); err != nil {
		logger.Errorf(ctx, "failed to delete slot: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "student %s cancelled sign up for queue %s", studentID, queueID)
	return nil
}

// ==================== Управление слотами (для старосты) ====================

// UpdateSlotStatus обновляет статус слота (сдал/не сдал/не явился)
func (q *QueueUseCase) UpdateSlotStatus(ctx context.Context, headmanID, slotID uuid.UUID, status models.SlotStatus) error {
	slot, err := q.queueRepo.GetSlotByID(ctx, slotID)
	if err != nil {
		return apperrors.ErrSlotNotFound
	}

	queue, err := q.queueRepo.GetByID(ctx, slot.QueueID)
	if err != nil {
		return apperrors.ErrQueueNotFound
	}

	if queue.CreatedByUserID != headmanID {
		return apperrors.ErrForbidden
	}

	slot.Status = status

	if err := q.queueRepo.UpdateSlot(ctx, slot); err != nil {
		logger.Errorf(ctx, "failed to update slot: %v", err)
		return apperrors.ErrInternalServer
	}

	logger.Infof(ctx, "slot %s status updated to %s", slotID, status)
	return nil
}

// MarkPassedCount помечает первых N человек как прошедших
func (q *QueueUseCase) MarkPassedCount(ctx context.Context, headmanID, queueID uuid.UUID, count int) error {
	queue, err := q.queueRepo.GetByID(ctx, queueID)
	if err != nil {
		return apperrors.ErrQueueNotFound
	}

	if queue.CreatedByUserID != headmanID {
		return apperrors.ErrForbidden
	}

	slots, err := q.queueRepo.GetSlotsByQueueID(ctx, queueID)
	if err != nil {
		logger.Errorf(ctx, "failed to get slots: %v", err)
		return apperrors.ErrInternalServer
	}

	// Помечаем первых count как passed, остальных как failed
	for i, slot := range slots {
		if slot.Status != models.SlotStatusWaiting {
			continue
		}

		if i < count {
			slot.Status = models.SlotStatusPassed
		} else {
			slot.Status = models.SlotStatusFailed
		}

		if err := q.queueRepo.UpdateSlot(ctx, slot); err != nil {
			logger.Errorf(ctx, "failed to update slot %s: %v", slot.ID, err)
		}
	}

	logger.Infof(ctx, "marked %d students as passed in queue %s", count, queueID)
	return nil
}

// TransferFailed переносит неуспевших из одной очереди в другую
func (q *QueueUseCase) TransferFailed(ctx context.Context, headmanID, fromQueueID, toQueueID uuid.UUID) (int, error) {
	toQueue, err := q.queueRepo.GetByID(ctx, toQueueID)
	if err != nil {
		return 0, apperrors.ErrQueueNotFound
	}

	if toQueue.CreatedByUserID != headmanID {
		return 0, apperrors.ErrForbidden
	}

	return q.transferFailedStudents(ctx, fromQueueID, toQueueID)
}

// ==================== Приватные методы ====================

func (q *QueueUseCase) transferFailedStudents(ctx context.Context, fromQueueID, toQueueID uuid.UUID) (int, error) {
	failedSlots, err := q.queueRepo.GetFailedSlotsByQueueID(ctx, fromQueueID)
	if err != nil {
		return 0, err
	}

	if len(failedSlots) == 0 {
		return 0, nil
	}

	// Получаем текущую последнюю позицию в новой очереди
	lastPosition, err := q.queueRepo.GetLastPosition(ctx, toQueueID)
	if err != nil {
		return 0, err
	}

	newSlots := make([]models.QueueSlot, 0, len(failedSlots))
	for i, oldSlot := range failedSlots {
		newSlots = append(newSlots, models.QueueSlot{
			ID:         uuid.New(),
			QueueID:    toQueueID,
			StudentID:  oldSlot.StudentID,
			Status:     models.SlotStatusWaiting,
			SignedUpAt: time.Now(),
		})
		_ = i
		_ = lastPosition // позиция определяется порядком вставки
	}

	count, err := q.queueRepo.CreateSlots(ctx, newSlots)
	if err != nil {
		return 0, err
	}

	return count, nil
}
