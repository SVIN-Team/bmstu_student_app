//go:build unit

package module

import (
	"context"
	"time"

	"stud_hub/internal/models"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) GetUserByID(ctx context.Context, id uuid.UUID) (models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user models.User) (uuid.UUID, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user models.User) (models.User, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(models.User), args.Error(1)
}

type MockTokenRepository struct {
	mock.Mock
}

func (m *MockTokenRepository) SaveRefreshToken(ctx context.Context, token models.RefreshToken) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockTokenRepository) GetRefreshToken(ctx context.Context, jti uuid.UUID) (models.RefreshToken, error) {
	args := m.Called(ctx, jti)
	return args.Get(0).(models.RefreshToken), args.Error(1)
}

func (m *MockTokenRepository) DeleteRefreshToken(ctx context.Context, jti uuid.UUID) error {
	args := m.Called(ctx, jti)
	return args.Error(0)
}

func (m *MockTokenRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

type MockGroupRepository struct {
	mock.Mock
}

func (m *MockGroupRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Group, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetByName(ctx context.Context, name string) (models.Group, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(models.Group), args.Error(1)
}

func (m *MockGroupRepository) GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockGroupRepository) GetAll(ctx context.Context) ([]models.Group, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Group), args.Error(1)
}

func (m *MockGroupRepository) Create(ctx context.Context, group models.Group) (uuid.UUID, error) {
	args := m.Called(ctx, group)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockGroupRepository) Update(ctx context.Context, group models.Group) error {
	args := m.Called(ctx, group)
	return args.Error(0)
}

func (m *MockGroupRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockGroupRepository) HasUsers(ctx context.Context, id uuid.UUID) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}

type MockSubjectRepository struct {
	mock.Mock
}

func (m *MockSubjectRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Subject, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Subject), args.Error(1)
}

func (m *MockSubjectRepository) GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

type MockTeacherRepository struct {
	mock.Mock
}

func (m *MockTeacherRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Teacher, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Teacher), args.Error(1)
}

func (m *MockTeacherRepository) GetOrCreateByFullName(ctx context.Context, lastName, firstName, patronymic string) (uuid.UUID, error) {
	args := m.Called(ctx, lastName, firstName, patronymic)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

type MockClassroomRepository struct {
	mock.Mock
}

func (m *MockClassroomRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Classroom, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Classroom), args.Error(1)
}

func (m *MockClassroomRepository) GetOrCreateByName(ctx context.Context, name string) (uuid.UUID, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

type MockQueueRepository struct {
	mock.Mock
}

func (m *MockQueueRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Queue, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Queue), args.Error(1)
}

func (m *MockQueueRepository) GetByGroupID(ctx context.Context, groupID uuid.UUID) ([]models.Queue, error) {
	args := m.Called(ctx, groupID)
	return args.Get(0).([]models.Queue), args.Error(1)
}

func (m *MockQueueRepository) GetByLessonID(ctx context.Context, lessonID uuid.UUID) (*models.Queue, error) {
	args := m.Called(ctx, lessonID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Queue), args.Error(1)
}

func (m *MockQueueRepository) GetActiveByGroupAndSubject(ctx context.Context, groupID, subjectID uuid.UUID) (*models.Queue, error) {
	args := m.Called(ctx, groupID, subjectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Queue), args.Error(1)
}

func (m *MockQueueRepository) Create(ctx context.Context, queue models.Queue) (uuid.UUID, error) {
	args := m.Called(ctx, queue)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockQueueRepository) Update(ctx context.Context, queue models.Queue) error {
	args := m.Called(ctx, queue)
	return args.Error(0)
}

func (m *MockQueueRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockQueueSlotsRepository struct {
	mock.Mock
}

func (m *MockQueueSlotsRepository) GetSlotByID(ctx context.Context, id uuid.UUID) (models.QueueSlot, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.QueueSlot), args.Error(1)
}

func (m *MockQueueSlotsRepository) GetSlotByQueueAndStudent(ctx context.Context, queueID, studentID uuid.UUID) (*models.QueueSlot, error) {
	args := m.Called(ctx, queueID, studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueueSlot), args.Error(1)
}

func (m *MockQueueSlotsRepository) GetSlotsByQueueID(ctx context.Context, queueID uuid.UUID) ([]models.QueueSlot, error) {
	args := m.Called(ctx, queueID)
	return args.Get(0).([]models.QueueSlot), args.Error(1)
}

func (m *MockQueueSlotsRepository) GetSlotsCount(ctx context.Context, queueID uuid.UUID) (int, error) {
	args := m.Called(ctx, queueID)
	return args.Int(0), args.Error(1)
}

func (m *MockQueueSlotsRepository) CreateSlot(ctx context.Context, slot models.QueueSlot) (*models.QueueSlot, error) {
	args := m.Called(ctx, slot)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.QueueSlot), args.Error(1)
}

func (m *MockQueueSlotsRepository) CreateSlots(ctx context.Context, slots []models.QueueSlot) (int, error) {
	args := m.Called(ctx, slots)
	return args.Int(0), args.Error(1)
}

func (m *MockQueueSlotsRepository) UpdateSlot(ctx context.Context, slot models.QueueSlot) error {
	args := m.Called(ctx, slot)
	return args.Error(0)
}

func (m *MockQueueSlotsRepository) DeleteSlot(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockQueueSlotsRepository) GetFailedSlotsByQueueID(ctx context.Context, queueID uuid.UUID) ([]models.QueueSlot, error) {
	args := m.Called(ctx, queueID)
	return args.Get(0).([]models.QueueSlot), args.Error(1)
}

func (m *MockQueueSlotsRepository) GetLastPosition(ctx context.Context, queueID uuid.UUID) (int, error) {
	args := m.Called(ctx, queueID)
	return args.Int(0), args.Error(1)
}

type MockLessonRepository struct {
	mock.Mock
}

func (m *MockLessonRepository) GetByID(ctx context.Context, id uuid.UUID) (models.Lesson, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.Lesson), args.Error(1)
}

func (m *MockLessonRepository) GetByGroupAndDateRange(ctx context.Context, groupID uuid.UUID, from, to time.Time) ([]models.Lesson, error) {
	args := m.Called(ctx, groupID, from, to)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Lesson), args.Error(1)
}

func (m *MockLessonRepository) Create(ctx context.Context, lesson models.Lesson) (uuid.UUID, error) {
	args := m.Called(ctx, lesson)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockLessonRepository) Update(ctx context.Context, lesson models.Lesson) error {
	args := m.Called(ctx, lesson)
	return args.Error(0)
}

func (m *MockLessonRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockLessonRepository) CheckOverlap(ctx context.Context, groupID uuid.UUID, startsAt, endsAt time.Time, excludeID *uuid.UUID) (bool, error) {
	args := m.Called(ctx, groupID, startsAt, endsAt, excludeID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLessonRepository) BulkCreate(ctx context.Context, lessons []models.Lesson) (int, error) {
	args := m.Called(ctx, lessons)
	return args.Int(0), args.Error(1)
}

func (m *MockLessonRepository) DeleteByGroupAndDateRange(ctx context.Context, groupID uuid.UUID, from, to time.Time) (int, error) {
	args := m.Called(ctx, groupID, from, to)
	return args.Int(0), args.Error(1)
}

func (m *MockLessonRepository) GetLessonDetails(ctx context.Context, id uuid.UUID) (models.LessonDetails, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(models.LessonDetails), args.Error(1)
}
