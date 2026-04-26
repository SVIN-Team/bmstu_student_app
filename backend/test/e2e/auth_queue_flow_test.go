//go:build e2e
// +build e2e

package e2e

import (
    "testing"
    "time"

    "stud_hub/internal/models"

    "github.com/stretchr/testify/require"
)

func TestAuthQueueFlow(t *testing.T) {
    services, cleanup := SetupTestServices(t)
    defer cleanup()
    
    ctx, cancel := WithTestContext(t)
    defer cancel()
    
    // Регистрация
    studentEmail := "e2e_student@test.com"
    studentPassword := "password123"
    
    student := models.User{
        Email:        studentEmail,
        PasswordHash: studentPassword,
        FirstName:    "E2E",
        LastName:     "Student",
        Role:         models.RoleHeadman,
    }
    
    _, _, studentID, err := services.AuthUC.SignUp(ctx, student)
    require.NoError(t, err)
    t.Logf("Student registered: %s", studentID)
    
    // Вход
    loginStudent := models.User{
        Email:        studentEmail,
        PasswordHash: studentPassword,
    }
    
    _, newRefreshToken, _, err := services.AuthUC.SignIn(ctx, loginStudent)
    require.NoError(t, err)
    t.Logf("Student logged in")
    
    // Refresh токена
    _, refreshedRefreshToken, err := services.AuthUC.Refresh(ctx, newRefreshToken)
    require.NoError(t, err)
    t.Logf("Tokens refreshed")
    
    // Создание группы
    group, err := CreateTestGroup(services.DB, "E2E Test Group")
    require.NoError(t, err)
    require.NotNil(t, group)
    t.Logf("Group created: %s", group.ID)
    
    // Обновление группы студента
    student.GroupID = group.ID
    student.ID = studentID
    _, err = services.UserUC.UpdateUser(ctx, student)
    require.NoError(t, err)
    t.Logf("Student group updated")
    
    // Проверка обновления
    updatedUser, err := services.UserUC.GetUserByID(ctx, studentID)
    require.NoError(t, err)
    require.Equal(t, group.ID, updatedUser.GroupID)
    
    // Создание предмета
    subject := CreateTestSubject(services.DB, "E2E Test Subject")
    require.NotNil(t, subject)
    t.Logf("Subject created: %s", subject.ID)
    
    // Создание преподавателя
    teacher := CreateTestTeacher(services.DB, "Teacher", "E2E")
    require.NotNil(t, teacher)
    t.Logf("Teacher created: %s", teacher.ID)
    
    // Создание аудитории
    classroom := CreateTestClassroom(services.DB, "301")
    require.NotNil(t, classroom)
    t.Logf("Classroom created: %s", classroom.ID)
    
    // Создание занятия
    now := time.Now()
    lessonStartsAt := time.Date(now.Year(), now.Month(), now.Day(), 10, 0, 0, 0, time.UTC)
    lessonEndsAt := lessonStartsAt.Add(1 * time.Hour)
    
    lesson := CreateTestLesson(services.DB, group.ID, subject.ID, teacher.ID, classroom.ID, lessonStartsAt, lessonEndsAt)
    require.NotNil(t, lesson)
    t.Logf("Lesson created: %s", lesson.ID)
    
    // Создание очереди
    queueOpensAt := time.Now().UTC().Add(-1 * time.Hour)
    queueParams := models.CreateQueueParams{
        LessonID:        lesson.ID,
        CreatedByUserID: studentID,
        OpensAt:         queueOpensAt,
        MaxSize:         nil,
        ClosesAt:        nil,
    }
    
    queueID, err := services.QueueUC.Create(ctx, queueParams)
    require.NoError(t, err)
    t.Logf("Queue created: %s", queueID)
    
    // Проверка статуса после создания
    queueAfterCreate, err := services.QueueUC.GetByID(ctx, queueID)
    require.NoError(t, err)
    require.Equal(t, models.QueueStatusDraft, queueAfterCreate.Status)
    
    // Открытие очереди
    err = services.QueueUC.Open(ctx, studentID, queueID)
    require.NoError(t, err)
    t.Logf("Queue opened")
    
    // Проверка статуса после открытия
    queueAfterOpen, err := services.QueueUC.GetByID(ctx, queueID)
    require.NoError(t, err)
    require.Equal(t, models.QueueStatusOpen, queueAfterOpen.Status)
    
    // Запись в очередь
    slot, err := services.QueueUC.SignUp(ctx, studentID, queueID)
    require.NoError(t, err)
    t.Logf("Student signed up, position: %d", slot.Position)
    
    // Проверка слота
    slotFromUC, err := services.QueueUC.GetSlotByID(ctx, slot.ID)
    require.NoError(t, err)
    require.Equal(t, models.SlotStatusWaiting, slotFromUC.Status)
    
    // Отметка сдачи
    err = services.QueueUC.MarkPassedCount(ctx, studentID, queueID, 1)
    require.NoError(t, err)
    t.Logf("Marked as passed")
    
    // Проверка статуса после отметки
    updatedSlot, err := services.QueueUC.GetSlotByID(ctx, slot.ID)
    require.NoError(t, err)
    require.Equal(t, models.SlotStatusPassed, updatedSlot.Status)
    
    // Проверка количества слотов
    allSlots, err := services.QueueUC.GetSlotsByQueueID(ctx, queueID)
    require.NoError(t, err)
    require.Equal(t, 1, len(allSlots))
    
    // Выход
    err = services.AuthUC.SignOut(ctx, refreshedRefreshToken)
    require.NoError(t, err)
    t.Logf("Student signed out")
    
    // Проверка инвалидации токена
    _, _, err = services.AuthUC.Refresh(ctx, refreshedRefreshToken)
    require.Error(t, err)
    
    t.Logf(" E2E test completed successfully!")
}