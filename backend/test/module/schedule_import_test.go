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
	"github.com/stretchr/testify/require"
)

func TestScheduleUseCase_ImportSchedule_CreatesMissingGroup(t *testing.T) {
	ctx := context.Background()

	groupRepo := new(MockGroupRepository)
	subjectRepo := new(MockSubjectRepository)
	teacherRepo := new(MockTeacherRepository)
	classroomRepo := new(MockClassroomRepository)
	lessonRepo := new(MockLessonRepository)
	queueRepo := new(MockQueueRepository)

	groupID := uuid.New()
	subjectID := uuid.New()
	teacherID := uuid.New()
	roomID := uuid.New()

	groupRepo.On("GetOrCreateByName", ctx, "ИУ7-81Б").Return(groupID, nil)
	subjectRepo.On("GetOrCreateByName", ctx, "Базы данных").Return(subjectID, nil)
	teacherRepo.On("GetOrCreateByFullName", ctx, "Иванов", "Иван", "Иванович").Return(teacherID, nil)
	classroomRepo.On("GetOrCreateByName", ctx, "ГУК-513").Return(roomID, nil)

	var capturedLessons []models.Lesson
	lessonRepo.
		On("DeleteByGroupAndDateRange", ctx, groupID, time.Date(2026, time.February, 23, 9, 0, 0, 0, time.UTC), time.Date(2026, time.February, 23, 10, 30, 0, 0, time.UTC)).
		Return(0, nil)
	lessonRepo.
		On("BulkCreate", ctx, mock.AnythingOfType("[]models.Lesson")).
		Run(func(args mock.Arguments) {
			capturedLessons = append([]models.Lesson(nil), args.Get(1).([]models.Lesson)...)
		}).
		Return(1, nil)

	scheduleUC := usecase.NewScheduleUseCase(lessonRepo, subjectRepo, groupRepo, teacherRepo, classroomRepo, queueRepo)

	rows := []models.ScheduleImportRow{{
		GroupName:         "ИУ7-81Б",
		SubjectName:       "Базы данных",
		TeacherLastName:   "Иванов",
		TeacherFirstName:  "Иван",
		TeacherPatronymic: "Иванович",
		RoomName:          "ГУК-513",
		LessonType:        models.LessonTypeLecture,
		StartsAt:          time.Date(2026, time.February, 23, 9, 0, 0, 0, time.UTC),
		EndsAt:            time.Date(2026, time.February, 23, 10, 30, 0, 0, time.UTC),
		LineNum:           1,
	}}

	result, err := scheduleUC.ImportSchedule(ctx, rows, true)

	require.NoError(t, err)
	assert.Equal(t, 1, result.TotalRows)
	assert.Equal(t, 1, result.SuccessCount)
	assert.Empty(t, result.Errors)
	require.Len(t, capturedLessons, 1)
	assert.Equal(t, groupID, capturedLessons[0].GroupID)
	assert.Equal(t, subjectID, capturedLessons[0].SubjectID)
	assert.Equal(t, teacherID, capturedLessons[0].TeacherID)
	assert.Equal(t, roomID, capturedLessons[0].RoomID)

	groupRepo.AssertExpectations(t)
	subjectRepo.AssertExpectations(t)
	teacherRepo.AssertExpectations(t)
	classroomRepo.AssertExpectations(t)
	assert.True(t, lessonRepo.AssertExpectations(t))
}
