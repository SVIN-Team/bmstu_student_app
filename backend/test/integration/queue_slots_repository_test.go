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
    "gorm.io/gorm"
)

func TestQueueSlotsRepository_CreateSlot(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Slot Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Slot Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_slot@example.com", "Creator", "Slot")
    require.NotNil(t, creator)

    student := CreateTestUser(db, "student@example.com", "Student", "User")
    require.NotNil(t, student)

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    slot := models.QueueSlot{
        ID:         uuid.New(),
        QueueID:    queue.ID,
        StudentID:  student.ID,
        Status:     models.SlotStatusWaiting,
        SignedUpAt: time.Now(),
    }

    createdSlot, err := repo.CreateSlot(ctx, slot)
    assert.NoError(t, err)
    assert.NotNil(t, createdSlot)
    assert.Equal(t, slot.ID, createdSlot.ID)
    assert.Equal(t, 1, createdSlot.Position)

    savedSlot, err := repo.GetSlotByID(ctx, slot.ID)
    assert.NoError(t, err)
    assert.Equal(t, slot.QueueID, savedSlot.QueueID)
    assert.Equal(t, slot.StudentID, savedSlot.StudentID)
}

func TestQueueSlotsRepository_GetSlotByID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Get Slot Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Get Slot Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_get@example.com", "Creator", "Get")
    require.NotNil(t, creator)

    student := CreateTestUser(db, "getslot@example.com", "Get", "Slot")
    require.NotNil(t, student)

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    testSlot := gormmodels.QueueSlot{
        ID:         uuid.New(),
        QueueID:    queue.ID,
        StudentID:  student.ID,
        Status:     gormmodels.QueueSlotStatusWaiting,
        SignedUpAt: time.Now(),
    }
    err := db.Create(&testSlot).Error
    require.NoError(t, err)

    slot, err := repo.GetSlotByID(ctx, testSlot.ID)
    assert.NoError(t, err)
    assert.Equal(t, testSlot.ID, slot.ID)
    assert.Equal(t, testSlot.QueueID, slot.QueueID)
    assert.Equal(t, testSlot.StudentID, slot.StudentID)
}

func TestQueueSlotsRepository_GetSlotByID_NotFound(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    nonExistentID := uuid.New()
    slot, err := repo.GetSlotByID(ctx, nonExistentID)

    assert.Error(t, err)
    assert.Equal(t, models.QueueSlot{}, slot)
}

func TestQueueSlotsRepository_GetSlotByQueueAndStudent(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Get Queue Student Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Get Queue Student Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_queue@example.com", "Creator", "Queue")
    require.NotNil(t, creator)

    student := CreateTestUser(db, "queuestudent@example.com", "Queue", "Student")
    require.NotNil(t, student)

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    slot := gormmodels.QueueSlot{
        ID:         uuid.New(),
        QueueID:    queue.ID,
        StudentID:  student.ID,
        Status:     gormmodels.QueueSlotStatusWaiting,
        SignedUpAt: time.Now(),
    }
    err := db.Create(&slot).Error
    require.NoError(t, err)

    foundSlot, err := repo.GetSlotByQueueAndStudent(ctx, queue.ID, student.ID)
    assert.NoError(t, err)
    assert.NotNil(t, foundSlot)
    assert.Equal(t, slot.ID, foundSlot.ID)
}

func TestQueueSlotsRepository_GetSlotByQueueAndStudent_NotFound(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    queueID := uuid.New()
    studentID := uuid.New()

    slot, err := repo.GetSlotByQueueAndStudent(ctx, queueID, studentID)
    assert.NoError(t, err)
    assert.Nil(t, slot)
}

func TestQueueSlotsRepository_GetSlotsByQueueID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "List Slots Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "List Slots Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_list@example.com", "Creator", "List")
    require.NotNil(t, creator)

    students := make([]*gormmodels.User, 3)
    for i := 0; i < 3; i++ {
        students[i] = CreateTestUser(db, "student_list_"+string(rune(65+i))+"@example.com", "Student", "List")
        require.NotNil(t, students[i])
    }

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    baseTime := time.Now()
    for i, student := range students {
        slot := gormmodels.QueueSlot{
            ID:         uuid.New(),
            QueueID:    queue.ID,
            StudentID:  student.ID,
            Status:     gormmodels.QueueSlotStatusWaiting,
            SignedUpAt: baseTime.Add(time.Duration(i) * time.Second),
        }
        err := db.Create(&slot).Error
        require.NoError(t, err)
    }

    slots, err := repo.GetSlotsByQueueID(ctx, queue.ID)
    assert.NoError(t, err)
    assert.Len(t, slots, 3)
}

func TestQueueSlotsRepository_GetSlotsCount(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Count Slots Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Count Slots Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_count@example.com", "Creator", "Count")
    require.NotNil(t, creator)

    students := make([]*gormmodels.User, 3)
    for i := 0; i < 3; i++ {
        students[i] = CreateTestUser(db, "count_student_"+string(rune(65+i))+"@example.com", "Count", "Student")
        require.NotNil(t, students[i])
    }

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    for i := 0; i < 3; i++ {
        slot := gormmodels.QueueSlot{
            ID:         uuid.New(),
            QueueID:    queue.ID,
            StudentID:  students[i].ID,
            Status:     gormmodels.QueueSlotStatusWaiting,
            SignedUpAt: time.Now(),
        }
        err := db.Create(&slot).Error
        require.NoError(t, err)
    }

    count, err := repo.GetSlotsCount(ctx, queue.ID)
    assert.NoError(t, err)
    assert.Equal(t, 3, count)
}

func TestQueueSlotsRepository_GetLastPosition(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Position Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Position Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_position@example.com", "Creator", "Position")
    require.NotNil(t, creator)

    students := make([]*gormmodels.User, 5)
    for i := 0; i < 5; i++ {
        students[i] = CreateTestUser(db, "position_student_"+string(rune(65+i))+"@example.com", "Position", "Student")
        require.NotNil(t, students[i])
    }

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    for i := 0; i < 5; i++ {
        slot := models.QueueSlot{
            ID:         uuid.New(),
            QueueID:    queue.ID,
            StudentID:  students[i].ID,
            Status:     models.SlotStatusWaiting,
            SignedUpAt: time.Now(),
        }
        _, err := repo.CreateSlot(ctx, slot)
        require.NoError(t, err)
    }

    lastPos, err := repo.GetLastPosition(ctx, queue.ID)
    assert.NoError(t, err)
    assert.Equal(t, 5, lastPos)
}

func TestQueueSlotsRepository_UpdateSlot(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Update Slot Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Update Slot Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_update@example.com", "Creator", "Update")
    require.NotNil(t, creator)

    student := CreateTestUser(db, "updateslot@example.com", "Update", "Slot")
    require.NotNil(t, student)

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    slot := models.QueueSlot{
        ID:         uuid.New(),
        QueueID:    queue.ID,
        StudentID:  student.ID,
        Status:     models.SlotStatusWaiting,
        SignedUpAt: time.Now(),
    }

    createdSlot, err := repo.CreateSlot(ctx, slot)
    require.NoError(t, err)

    createdSlot.Status = models.SlotStatusPassed
    err = repo.UpdateSlot(ctx, *createdSlot)
    assert.NoError(t, err)

    updatedSlot, err := repo.GetSlotByID(ctx, createdSlot.ID)
    assert.NoError(t, err)
    assert.Equal(t, models.SlotStatusPassed, updatedSlot.Status)
}

func TestQueueSlotsRepository_DeleteSlot(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Delete Slot Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Delete Slot Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_delete@example.com", "Creator", "Delete")
    require.NotNil(t, creator)

    student := CreateTestUser(db, "deleteslot@example.com", "Delete", "Slot")
    require.NotNil(t, student)

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    slot := models.QueueSlot{
        ID:         uuid.New(),
        QueueID:    queue.ID,
        StudentID:  student.ID,
        Status:     models.SlotStatusWaiting,
        SignedUpAt: time.Now(),
    }

    createdSlot, err := repo.CreateSlot(ctx, slot)
    require.NoError(t, err)

    err = repo.DeleteSlot(ctx, createdSlot.ID)
    assert.NoError(t, err)

    _, err = repo.GetSlotByID(ctx, createdSlot.ID)
    assert.Error(t, err)
}

func TestQueueSlotsRepository_CreateSlots(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Bulk Slots Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Bulk Slots Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_bulk@example.com", "Creator", "Bulk")
    require.NotNil(t, creator)

    students := make([]*gormmodels.User, 3)
    for i := 0; i < 3; i++ {
        students[i] = CreateTestUser(db, "bulk_"+string(rune(65+i))+"@example.com", "Bulk", "Slot")
        require.NotNil(t, students[i])
    }

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    slots := make([]models.QueueSlot, 3)
    for i, student := range students {
        slots[i] = models.QueueSlot{
            ID:         uuid.New(),
            QueueID:    queue.ID,
            StudentID:  student.ID,
            Status:     models.SlotStatusWaiting,
            SignedUpAt: time.Now(),
        }
    }

    count, err := repo.CreateSlots(ctx, slots)
    assert.NoError(t, err)
    assert.Equal(t, 3, count)

    savedSlots, err := repo.GetSlotsByQueueID(ctx, queue.ID)
    assert.NoError(t, err)
    assert.Len(t, savedSlots, 3)
}

func TestQueueSlotsRepository_GetFailedSlotsByQueueID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewQueueSlotsRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Failed Slots Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Failed Slots Subject")
    require.NotNil(t, subject)

    creator := CreateTestUser(db, "creator_failed@example.com", "Creator", "Failed")
    require.NotNil(t, creator)

    students := make([]*gormmodels.User, 3)
    for i := 0; i < 3; i++ {
        students[i] = CreateTestUser(db, "failed_"+string(rune(65+i))+"@example.com", "Failed", "Slot")
        require.NotNil(t, students[i])
    }

    queue := createTestQueue(t, db, group.ID, subject.ID, creator.ID)
    require.NotNil(t, queue)

    statuses := []gormmodels.QueueSlotStatus{
        gormmodels.QueueSlotStatusFailed,
        gormmodels.QueueSlotStatusWaiting,
        gormmodels.QueueSlotStatusNoShow,
    }

    for i, student := range students {
        slot := gormmodels.QueueSlot{
            ID:         uuid.New(),
            QueueID:    queue.ID,
            StudentID:  student.ID,
            Status:     statuses[i],
            SignedUpAt: time.Now(),
        }
        err := db.Create(&slot).Error
        require.NoError(t, err)
    }

    failedSlots, err := repo.GetFailedSlotsByQueueID(ctx, queue.ID)
    assert.NoError(t, err)
    assert.Len(t, failedSlots, 2)
}

func createTestQueue(t *testing.T, db *gorm.DB, groupID, subjectID, userID uuid.UUID) *gormmodels.Queue {
    queue := &gormmodels.Queue{
        ID:        uuid.New(),
        GroupID:   groupID,
        SubjectID: subjectID,
        CreatedBy: userID,
        OpensAt:   time.Now(),
        Status:    gormmodels.QueueStatusOpen,
    }
    err := db.Create(queue).Error
    require.NoError(t, err)
    return queue
}