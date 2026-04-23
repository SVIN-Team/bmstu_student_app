//go:build integration
// +build integration

package integration

import (
    "context"
    "testing"
    "time"

    "stud_hub/internal/models"
    gormmodels "stud_hub/internal/repository/gorm-models"
    gormrepository "stud_hub/internal/repository/gorm-repository"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestQueueRepository_Create(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Queue Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Queue Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "queue_creator@example.com", "Queue", "Creator")
    require.NotNil(t, creator)

    queue := models.Queue{
        ID:              uuid.New(),
        GroupID:         group.ID,
        SubjectID:       subject.ID,
        CreatedByUserID: creator.ID,
        CreatedAt:       time.Now(),
        OpensAt:         time.Now().Add(24 * time.Hour),
        Status:          models.QueueStatusDraft,
    }

    queueID, err := repo.Create(ctx, queue)
    assert.NoError(t, err)
    assert.Equal(t, queue.ID, queueID)

    savedQueue, err := repo.GetByID(ctx, queueID)
    assert.NoError(t, err)
    assert.Equal(t, queue.GroupID, savedQueue.GroupID)
    assert.Equal(t, queue.SubjectID, savedQueue.SubjectID)
    assert.Equal(t, queue.CreatedByUserID, savedQueue.CreatedByUserID)
}

func TestQueueRepository_GetByID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Get Queue Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Get Queue Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "get_queue_creator@example.com", "Get", "Queue")
    require.NotNil(t, creator)

    testQueue := gormmodels.Queue{
        ID:        uuid.New(),
        GroupID:   group.ID,
        SubjectID: subject.ID,
        CreatedBy: creator.ID,
        CreatedAt: time.Now(),
        OpensAt:   time.Now().Add(24 * time.Hour),
        Status:    gormmodels.QueueStatusDraft,
    }
    err := db.Create(&testQueue).Error
    require.NoError(t, err)

    queue, err := repo.GetByID(ctx, testQueue.ID)
    assert.NoError(t, err)
    assert.Equal(t, testQueue.ID, queue.ID)
    assert.Equal(t, testQueue.GroupID, queue.GroupID)
    assert.Equal(t, testQueue.SubjectID, queue.SubjectID)
    assert.Equal(t, testQueue.CreatedBy, queue.CreatedByUserID)
}

func TestQueueRepository_GetByID_NotFound(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    nonExistentID := uuid.New()
    queue, err := repo.GetByID(ctx, nonExistentID)

    assert.Error(t, err)
    assert.Equal(t, models.Queue{}, queue)
}

func TestQueueRepository_GetByGroupID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Group Queues")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Group Queues Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "group_queues@example.com", "Group", "Queues")
    require.NotNil(t, creator)

    for i := 0; i < 3; i++ {
        queue := gormmodels.Queue{
            ID:        uuid.New(),
            GroupID:   group.ID,
            SubjectID: subject.ID,
            CreatedBy: creator.ID,
            CreatedAt: time.Now(),
            OpensAt:   time.Now().Add(time.Duration(i+1) * 24 * time.Hour),
            Status:    gormmodels.QueueStatusDraft,
        }
        err := db.Create(&queue).Error
        require.NoError(t, err)
    }

    queues, err := repo.GetByGroupID(ctx, group.ID)
    assert.NoError(t, err)
    assert.Len(t, queues, 3)
}

func TestQueueRepository_GetByLessonID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Lesson Queue Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Lesson Queue Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Lesson", "Teacher")
    require.NotNil(t, teacher)

    creator := CreateTestUser(db, "lesson_queue@example.com", "Lesson", "Queue")
    require.NotNil(t, creator)

    lesson := gormmodels.Lesson{
        ID:        uuid.New(),
        GroupID:   group.ID,
        SubjectID: subject.ID,
        TeacherID: &teacher.ID,
        Type:      gormmodels.LessonTypeLecture,
        StartsAt:  time.Now().Add(48 * time.Hour),
        EndsAt:    time.Now().Add(50 * time.Hour),
    }
    err := db.Create(&lesson).Error
    require.NoError(t, err)

    testQueue := gormmodels.Queue{
        ID:        uuid.New(),
        GroupID:   group.ID,
        SubjectID: subject.ID,
        LessonID:  &lesson.ID,
        CreatedBy: creator.ID,
        CreatedAt: time.Now(),
        OpensAt:   time.Now().Add(24 * time.Hour),
        Status:    gormmodels.QueueStatusOpen,
    }
    err = db.Create(&testQueue).Error
    require.NoError(t, err)

    queue, err := repo.GetByLessonID(ctx, lesson.ID)
    assert.NoError(t, err)
    assert.NotNil(t, queue)
    assert.Equal(t, testQueue.ID, queue.ID)
}

func TestQueueRepository_GetByLessonID_NotFound(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    nonExistentID := uuid.New()
    queue, err := repo.GetByLessonID(ctx, nonExistentID)

    assert.Error(t, err)
    assert.Nil(t, queue)
}

func TestQueueRepository_GetActiveByGroupAndSubject(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Active Queue Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Active Queue Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "active_queue@example.com", "Active", "Queue")
    require.NotNil(t, creator)

    activeQueue := gormmodels.Queue{
        ID:        uuid.New(),
        GroupID:   group.ID,
        SubjectID: subject.ID,
        CreatedBy: creator.ID,
        CreatedAt: time.Now(),
        OpensAt:   time.Now().Add(-1 * time.Hour), 
        Status:    gormmodels.QueueStatusOpen,
    }
    err := db.Create(&activeQueue).Error
    require.NoError(t, err)

    draftQueue := gormmodels.Queue{
        ID:        uuid.New(),
        GroupID:   group.ID,
        SubjectID: subject.ID,
        CreatedBy: creator.ID,
        CreatedAt: time.Now(),
        OpensAt:   time.Now().Add(24 * time.Hour),
        Status:    gormmodels.QueueStatusDraft,
    }
    err = db.Create(&draftQueue).Error
    require.NoError(t, err)

    queue, err := repo.GetActiveByGroupAndSubject(ctx, group.ID, subject.ID)
    assert.NoError(t, err)
    assert.NotNil(t, queue)
    assert.Equal(t, activeQueue.ID, queue.ID)
    assert.Equal(t, models.QueueStatusOpen, queue.Status)
}

func TestQueueRepository_GetActiveByGroupAndSubject_NotFound(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "No Active Queue Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "No Active Queue Subject")
    require.NotNil(t, subject)

    queue, err := repo.GetActiveByGroupAndSubject(ctx, group.ID, subject.ID)
    assert.Error(t, err)
    assert.Nil(t, queue)
}

func TestQueueRepository_Update(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Update Queue Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Update Queue Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "update_queue@example.com", "Update", "Queue")
    require.NotNil(t, creator)

    opensAt := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)
    queue := models.Queue{
        ID:              uuid.New(),
        GroupID:         group.ID,
        SubjectID:       subject.ID,
        CreatedByUserID: creator.ID,
        CreatedAt:       time.Now().UTC().Truncate(time.Second),
        OpensAt:         opensAt,
        Status:          models.QueueStatusDraft,
    }

    queueID, err := repo.Create(ctx, queue)
    require.NoError(t, err)

    newOpensAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)
    queue.OpensAt = newOpensAt
    queue.Status = models.QueueStatusOpen

    err = repo.Update(ctx, queue)
    assert.NoError(t, err)

    updatedQueue, err := repo.GetByID(ctx, queueID)
    assert.NoError(t, err)
    assert.Equal(t, models.QueueStatusOpen, updatedQueue.Status)
    assert.True(t, updatedQueue.OpensAt.UTC().Truncate(time.Second).Equal(newOpensAt.UTC().Truncate(time.Second)))
}

func TestQueueRepository_CreateWithMaxSize(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "MaxSize Queue Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "MaxSize Queue Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "maxsize_queue@example.com", "MaxSize", "Queue")
    require.NotNil(t, creator)

    maxSize := uint32(10)
    closesAt := time.Now().UTC().Add(48 * time.Hour).Truncate(time.Second)

    queue := models.Queue{
        ID:              uuid.New(),
        GroupID:         group.ID,
        SubjectID:       subject.ID,
        CreatedByUserID: creator.ID,
        CreatedAt:       time.Now().UTC().Truncate(time.Second),
        OpensAt:         time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second),
        ClosesAt:        &closesAt,
        MaxSize:         &maxSize,
        Status:          models.QueueStatusOpen,
    }

    queueID, err := repo.Create(ctx, queue)
    assert.NoError(t, err)

    savedQueue, err := repo.GetByID(ctx, queueID)
    assert.NoError(t, err)
    assert.NotNil(t, savedQueue.MaxSize)
    assert.Equal(t, maxSize, *savedQueue.MaxSize)
    assert.NotNil(t, savedQueue.ClosesAt)
    assert.True(t, savedQueue.ClosesAt.UTC().Truncate(time.Second).Equal(closesAt.UTC().Truncate(time.Second)))
}

func TestQueueRepository_Delete(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Delete Queue Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Delete Queue Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "delete_queue@example.com", "Delete", "Queue")
    require.NotNil(t, creator)

    queue := models.Queue{
        ID:              uuid.New(),
        GroupID:         group.ID,
        SubjectID:       subject.ID,
        CreatedByUserID: creator.ID,
        CreatedAt:       time.Now(),
        OpensAt:         time.Now().Add(24 * time.Hour),
        Status:          models.QueueStatusDraft,
    }

    queueID, err := repo.Create(ctx, queue)
    require.NoError(t, err)

    err = repo.Delete(ctx, queueID)
    assert.NoError(t, err)

    _, err = repo.GetByID(ctx, queueID)
    assert.Error(t, err)
}
