//go:build unit

package module

import (
	"context"
	"testing"
	"time"

	"stud_hub/internal/models"
	"stud_hub/internal/usecase"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestQueueUseCase_SignUp_Success(t *testing.T) {
	ctx := context.Background()
	studentID := uuid.New()
	queueID := uuid.New()
	groupID := uuid.New()

	queue := models.Queue{
		ID:       queueID,
		GroupID:  groupID,
		Status:   models.QueueStatusOpen,
		OpensAt:  time.Now().Add(-1 * time.Hour),
		ClosesAt: nil,
		MaxSize:  nil,
	}

	user := models.User{
		ID:      studentID,
		GroupID: groupID,
	}

	expectedSlot := &models.QueueSlot{
		ID:        uuid.New(),
		QueueID:   queueID,
		StudentID: studentID,
		Status:    models.SlotStatusWaiting,
		Position:  1,
	}

	mockQueueRepo := new(MockQueueRepository)
	mockQueueSlotsRepo := new(MockQueueSlotsRepository)
	mockUserRepo := new(MockUserRepository)
	mockLessonRepo := new(MockLessonRepository)

	mockQueueRepo.On("GetByID", ctx, queueID).Return(queue, nil)
	mockUserRepo.On("GetUserByID", ctx, studentID).Return(user, nil)
	mockQueueSlotsRepo.On("GetSlotByQueueAndStudent", ctx, queueID, studentID).Return(nil, nil)
	mockQueueSlotsRepo.On("CreateSlot", ctx, mock.AnythingOfType("models.QueueSlot")).Return(expectedSlot, nil)

	qUC := usecase.NewQueueUseCase(mockQueueRepo, mockQueueSlotsRepo, mockUserRepo, mockLessonRepo)

	slot, err := qUC.SignUp(ctx, studentID, queueID)

	assert.NoError(t, err)
	assert.NotNil(t, slot)
	assert.Equal(t, studentID, slot.StudentID)

	mockQueueRepo.AssertExpectations(t)
    mockQueueSlotsRepo.AssertExpectations(t)
    mockUserRepo.AssertExpectations(t)
    mockLessonRepo.AssertExpectations(t)
}

func TestQueueUseCase_MarkPassedCount_Success(t *testing.T) {
	ctx := context.Background()
	headmanID := uuid.New()
	queueID := uuid.New()

	queue := models.Queue{
		ID:              queueID,
		CreatedByUserID: headmanID,
		Status:          models.QueueStatusOpen,
	}

	slots := []models.QueueSlot{
		{ID: uuid.New(), QueueID: queueID, Position: 1, Status: models.SlotStatusWaiting},
		{ID: uuid.New(), QueueID: queueID, Position: 2, Status: models.SlotStatusWaiting},
		{ID: uuid.New(), QueueID: queueID, Position: 3, Status: models.SlotStatusWaiting},
	}

	mockQueueRepo := new(MockQueueRepository)
	mockQueueSlotsRepo := new(MockQueueSlotsRepository)
	mockUserRepo := new(MockUserRepository)
	mockLessonRepo := new(MockLessonRepository)

	mockQueueRepo.On("GetByID", ctx, queueID).Return(queue, nil)
	mockQueueSlotsRepo.On("GetSlotsByQueueID", ctx, queueID).Return(slots, nil)
	mockQueueSlotsRepo.On("UpdateSlot", ctx, mock.AnythingOfType("models.QueueSlot")).Return(nil).Times(3)

	qUC := usecase.NewQueueUseCase(mockQueueRepo, mockQueueSlotsRepo, mockUserRepo, mockLessonRepo)

	err := qUC.MarkPassedCount(ctx, headmanID, queueID, 2)

	assert.NoError(t, err)
}