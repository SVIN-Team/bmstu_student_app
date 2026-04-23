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

func TestLessonRepository_Create(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()
    group := CreateTestGroup(db, "Test Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Test Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Test", "Teacher")
    require.NotNil(t, teacher)

    startsAt := time.Now().Add(24 * time.Hour)
    endsAt := startsAt.Add(2 * time.Hour)

    lesson := models.Lesson{
        ID:         uuid.New(),
        GroupID:    group.ID,
        SubjectID:  subject.ID,
        TeacherID:  teacher.ID,
        LessonType: models.LessonTypeLecture,
        StartsAt:   startsAt,
        EndsAt:     endsAt,
    }

    lessonID, err := repo.Create(ctx, lesson)

    assert.NoError(t, err)
    assert.Equal(t, lesson.ID, lessonID)

    savedLesson, err := repo.GetByID(ctx, lessonID)
    assert.NoError(t, err)
    assert.Equal(t, lesson.GroupID, savedLesson.GroupID)
    assert.Equal(t, lesson.SubjectID, savedLesson.SubjectID)
    assert.Equal(t, lesson.TeacherID, savedLesson.TeacherID)
    assert.Equal(t, lesson.LessonType, savedLesson.LessonType)
}

func TestLessonRepository_GetByID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Get Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Get Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Get", "Teacher")
    require.NotNil(t, teacher)

    startsAt := time.Now().Add(48 * time.Hour)
    endsAt := startsAt.Add(2 * time.Hour)

    testLesson := gormmodels.Lesson{
        ID:        uuid.New(),
        GroupID:   group.ID,
        SubjectID: subject.ID,
        TeacherID: &teacher.ID,
        Type:      gormmodels.LessonTypeLecture,
        StartsAt:  startsAt,
        EndsAt:    endsAt,
    }
    err := db.Create(&testLesson).Error
    require.NoError(t, err)

    lesson, err := repo.GetByID(ctx, testLesson.ID)
    assert.NoError(t, err)
    assert.Equal(t, testLesson.ID, lesson.ID)
    assert.Equal(t, testLesson.GroupID, lesson.GroupID)
    assert.Equal(t, testLesson.SubjectID, lesson.SubjectID)
    assert.Equal(t, testLesson.TeacherID, lesson.TeacherID)
}

func TestLessonRepository_GetByID_NotFound(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()

    nonExistentID := uuid.New()
    lesson, err := repo.GetByID(ctx, nonExistentID)

    assert.Error(t, err)
    assert.Equal(t, models.Lesson{}, lesson)
}

func TestLessonRepository_Update(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()
    group := CreateTestGroup(db, "Update Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Update Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Update", "Teacher")
    require.NotNil(t, teacher)
    startsAt := time.Now().UTC().Add(72 * time.Hour).Truncate(time.Second)
    endsAt := startsAt.Add(2 * time.Hour)

    lesson := models.Lesson{
        ID:         uuid.New(),
        GroupID:    group.ID,
        SubjectID:  subject.ID,
        TeacherID:  teacher.ID,
        LessonType: models.LessonTypeLecture,
        StartsAt:   startsAt,
        EndsAt:     endsAt,
    }

    lessonID, err := repo.Create(ctx, lesson)
    require.NoError(t, err)
    newStartsAt := startsAt.Add(24 * time.Hour)
    newEndsAt := newStartsAt.Add(2 * time.Hour)

    updatedLesson := models.Lesson{
        ID:         lessonID,
        GroupID:    group.ID,
        SubjectID:  subject.ID,
        TeacherID:  teacher.ID,
        LessonType: models.LessonTypeLab,
        StartsAt:   newStartsAt,
        EndsAt:     newEndsAt,
    }

    err = repo.Update(ctx, updatedLesson)
    assert.NoError(t, err)
    savedLesson, err := repo.GetByID(ctx, lessonID)
    assert.NoError(t, err)
    assert.Equal(t, models.LessonTypeLab, savedLesson.LessonType)
    assert.True(t, savedLesson.StartsAt.UTC().Truncate(time.Second).Equal(newStartsAt.UTC().Truncate(time.Second)))
    assert.True(t, savedLesson.EndsAt.UTC().Truncate(time.Second).Equal(newEndsAt.UTC().Truncate(time.Second)))
}

func TestLessonRepository_Delete(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()
    group := CreateTestGroup(db, "Delete Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Delete Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Delete", "Teacher")
    require.NotNil(t, teacher)
    startsAt := time.Now().Add(96 * time.Hour)
    endsAt := startsAt.Add(2 * time.Hour)

    lesson := models.Lesson{
        ID:         uuid.New(),
        GroupID:    group.ID,
        SubjectID:  subject.ID,
        TeacherID:  teacher.ID,
        LessonType: models.LessonTypeLecture,
        StartsAt:   startsAt,
        EndsAt:     endsAt,
    }

    lessonID, err := repo.Create(ctx, lesson)
    require.NoError(t, err)
    err = repo.Delete(ctx, lessonID)
    assert.NoError(t, err)
    _, err = repo.GetByID(ctx, lessonID)
    assert.Error(t, err)
}

func TestLessonRepository_GetByGroupAndDateRange(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()
    group := CreateTestGroup(db, "Range Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Range Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Range", "Teacher")
    require.NotNil(t, teacher)
    baseTime := time.Now().Add(120 * time.Hour)

    lessons := []models.Lesson{
        {
            ID:         uuid.New(),
            GroupID:    group.ID,
            SubjectID:  subject.ID,
            TeacherID:  teacher.ID,
            LessonType: models.LessonTypeLecture,
            StartsAt:   baseTime,
            EndsAt:     baseTime.Add(2 * time.Hour),
        },
        {
            ID:         uuid.New(),
            GroupID:    group.ID,
            SubjectID:  subject.ID,
            TeacherID:  teacher.ID,
            LessonType: models.LessonTypeLab,
            StartsAt:   baseTime.Add(24 * time.Hour),
            EndsAt:     baseTime.Add(26 * time.Hour),
        },
        {
            ID:         uuid.New(),
            GroupID:    group.ID,
            SubjectID:  subject.ID,
            TeacherID:  teacher.ID,
            LessonType: models.LessonTypeSeminar,
            StartsAt:   baseTime.Add(48 * time.Hour),
            EndsAt:     baseTime.Add(50 * time.Hour),
        },
    }

    for _, l := range lessons {
        _, err := repo.Create(ctx, l)
        require.NoError(t, err)
    }
    from := baseTime.Add(-1 * time.Hour)
    to := baseTime.Add(72 * time.Hour)

    result, err := repo.GetByGroupAndDateRange(ctx, group.ID, from, to)
    assert.NoError(t, err)
    assert.Len(t, result, 3)
}

func TestLessonRepository_CheckOverlap(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Overlap Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Overlap Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Overlap", "Teacher")
    require.NotNil(t, teacher)

    startsAt := time.Now().Add(144 * time.Hour)
    endsAt := startsAt.Add(2 * time.Hour)

    lesson := models.Lesson{
        ID:         uuid.New(),
        GroupID:    group.ID,
        SubjectID:  subject.ID,
        TeacherID:  teacher.ID,
        LessonType: models.LessonTypeLecture,
        StartsAt:   startsAt,
        EndsAt:     endsAt,
    }

    _, err := repo.Create(ctx, lesson)
    require.NoError(t, err)

    newStartsAt := startsAt.Add(30 * time.Minute)
    newEndsAt := endsAt.Add(30 * time.Minute)

    hasOverlap, err := repo.CheckOverlap(ctx, group.ID, newStartsAt, newEndsAt, nil)
    assert.NoError(t, err)
    assert.True(t, hasOverlap)

    noOverlapStartsAt := endsAt.Add(1 * time.Hour)
    noOverlapEndsAt := noOverlapStartsAt.Add(2 * time.Hour)

    hasOverlap, err = repo.CheckOverlap(ctx, group.ID, noOverlapStartsAt, noOverlapEndsAt, nil)
    assert.NoError(t, err)
    assert.False(t, hasOverlap)
}

func TestLessonRepository_BulkCreate(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Bulk Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Bulk Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Bulk", "Teacher")
    require.NotNil(t, teacher)

    baseTime := time.Now().Add(168 * time.Hour)
    lessons := make([]models.Lesson, 5)

    for i := 0; i < 5; i++ {
        lessons[i] = models.Lesson{
            ID:         uuid.New(),
            GroupID:    group.ID,
            SubjectID:  subject.ID,
            TeacherID:  teacher.ID,
            LessonType: models.LessonTypeLecture,
            StartsAt:   baseTime.Add(time.Duration(i*24) * time.Hour),
            EndsAt:     baseTime.Add(time.Duration(i*24+2) * time.Hour),
        }
    }

    count, err := repo.BulkCreate(ctx, lessons)
    assert.NoError(t, err)
    assert.Equal(t, 5, count)

    from := baseTime.Add(-1 * time.Hour)
    to := baseTime.Add(120 * time.Hour)

    result, err := repo.GetByGroupAndDateRange(ctx, group.ID, from, to)
    assert.NoError(t, err)
    assert.Len(t, result, 5)
}

func TestLessonRepository_DeleteByGroupAndDateRange(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Delete Range Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Delete Range Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Delete", "Teacher")
    require.NotNil(t, teacher)

    baseTime := time.Now().Add(192 * time.Hour)

    lessons := []models.Lesson{
        {
            ID:         uuid.New(),
            GroupID:    group.ID,
            SubjectID:  subject.ID,
            TeacherID:  teacher.ID,
            LessonType: models.LessonTypeLecture,
            StartsAt:   baseTime,
            EndsAt:     baseTime.Add(2 * time.Hour),
        },
        {
            ID:         uuid.New(),
            GroupID:    group.ID,
            SubjectID:  subject.ID,
            TeacherID:  teacher.ID,
            LessonType: models.LessonTypeLab,
            StartsAt:   baseTime.Add(24 * time.Hour),
            EndsAt:     baseTime.Add(26 * time.Hour),
        },
    }

    for _, l := range lessons {
        _, err := repo.Create(ctx, l)
        require.NoError(t, err)
    }

    from := baseTime.Add(-1 * time.Hour)
    to := baseTime.Add(48 * time.Hour)

    deleted, err := repo.DeleteByGroupAndDateRange(ctx, group.ID, from, to)
    assert.NoError(t, err)
    assert.Equal(t, 2, deleted)

    result, err := repo.GetByGroupAndDateRange(ctx, group.ID, from, to)
    assert.NoError(t, err)
    assert.Len(t, result, 0)
}

func TestLessonRepository_GetLessonDetails(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewLessonRepository(db)
    ctx := context.Background()

    group := CreateTestGroup(db, "Details Group")
    require.NotNil(t, group)

    subject := CreateTestSubject(db, "Details Subject")
    require.NotNil(t, subject)

    teacher := CreateTestTeacher(db, "Details", "Teacher")
    require.NotNil(t, teacher)

    startsAt := time.Now().Add(216 * time.Hour)
    endsAt := startsAt.Add(2 * time.Hour)

    lesson := models.Lesson{
        ID:         uuid.New(),
        GroupID:    group.ID,
        SubjectID:  subject.ID,
        TeacherID:  teacher.ID,
        LessonType: models.LessonTypeLecture,
        StartsAt:   startsAt,
        EndsAt:     endsAt,
    }

    lessonID, err := repo.Create(ctx, lesson)
    require.NoError(t, err)

    details, err := repo.GetLessonDetails(ctx, lessonID)
    assert.NoError(t, err)
    assert.Equal(t, lessonID, details.ID)
    assert.Equal(t, group.Name, details.GroupName)
    assert.Equal(t, subject.Name, details.SubjectName)
    assert.Equal(t, teacher.Name(), details.TeacherName)
}
