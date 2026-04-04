//go:build integration
// +build integration

package integration

import (
    "context"
    "testing"

    "stud_hub/internal/models"
    gormmodels "stud_hub/internal/repository/gorm-models"
    gormrepository "stud_hub/internal/repository/gorm-repository"

    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestClassroomRepository_GetByID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()

    roomName := "Test Room 101"
    testRoom := &gormmodels.Room{
        Name: &roomName,
    }
    err := db.Create(testRoom).Error
    require.NoError(t, err)

    classroom, err := repo.GetByID(ctx, testRoom.ID)

    assert.NoError(t, err)
    assert.Equal(t, testRoom.ID, classroom.ID)
    assert.Equal(t, roomName, classroom.Name)
}

func TestClassroomRepository_GetByID_NotFound(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()

    nonExistentID := uuid.New()
    classroom, err := repo.GetByID(ctx, nonExistentID)

    assert.Error(t, err)
    assert.Equal(t, models.Classroom{}, classroom)
}

func TestClassroomRepository_GetByID_WithNilName(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()

    testRoom := &gormmodels.Room{
        Name: nil,
    }
    err := db.Create(testRoom).Error
    require.NoError(t, err)

    classroom, err := repo.GetByID(ctx, testRoom.ID)

    assert.NoError(t, err)
    assert.Equal(t, testRoom.ID, classroom.ID)
    assert.Equal(t, "", classroom.Name)
}

func TestClassroomRepository_GetOrCreateByName_Existing(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()

    roomName := "Existing Room 202"
    existingRoom := &gormmodels.Room{
        Name: &roomName,
    }
    err := db.Create(existingRoom).Error
    require.NoError(t, err)

    roomID, err := repo.GetOrCreateByName(ctx, roomName)

    assert.NoError(t, err)
    assert.Equal(t, existingRoom.ID, roomID)

    var count int64
    db.Model(&gormmodels.Room{}).Where("name = ?", roomName).Count(&count)
    assert.Equal(t, int64(1), count)
}

func TestClassroomRepository_GetOrCreateByName_New(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()

    roomName := "Brand New Room 303"

    roomID, err := repo.GetOrCreateByName(ctx, roomName)

    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, roomID)

    var savedRoom gormmodels.Room
    err = db.First(&savedRoom, "id = ?", roomID).Error
    assert.NoError(t, err)
    assert.Equal(t, roomName, *savedRoom.Name)
}

func TestClassroomRepository_GetOrCreateByName_CaseSensitive(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()
    lowerName := "room 404"
    existingRoom := &gormmodels.Room{
        Name: &lowerName,
    }
    err := db.Create(existingRoom).Error
    require.NoError(t, err)

    upperName := "ROOM 404"
    roomID, err := repo.GetOrCreateByName(ctx, upperName)
    assert.NoError(t, err)
    assert.NotEqual(t, existingRoom.ID, roomID)

    var count int64
    db.Model(&gormmodels.Room{}).Count(&count)
    assert.Equal(t, int64(2), count)
}

func TestClassroomRepository_GetOrCreateByName_Concurrent(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()

    roomName := "Concurrent Room 505"
    const goroutines = 5
    results := make(chan uuid.UUID, goroutines)
    errors := make(chan error, goroutines)

    for i := 0; i < goroutines; i++ {
        go func() {
            id, err := repo.GetOrCreateByName(ctx, roomName)
            if err != nil {
                errors <- err
            } else {
                results <- id
            }
        }()
    }

    var ids []uuid.UUID
    for i := 0; i < goroutines; i++ {
        select {
        case id := <-results:
            ids = append(ids, id)
        case err := <-errors:
            t.Errorf("Error in concurrent call: %v", err)
        }
    }

    assert.NotEmpty(t, ids)
    firstID := ids[0]
    for _, id := range ids {
        assert.Equal(t, firstID, id)
    }
    var count int64
    db.Model(&gormmodels.Room{}).Where("name = ?", roomName).Count(&count)
    assert.Equal(t, int64(1), count)
}

func TestClassroomRepository_GetOrCreateByName_MultipleDifferent(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()

    names := []string{"Room 601", "Room 602", "Room 603", "Room 604", "Room 605"}
    ids := make([]uuid.UUID, 0, len(names))

    for _, name := range names {
        id, err := repo.GetOrCreateByName(ctx, name)
        assert.NoError(t, err)
        assert.NotEqual(t, uuid.Nil, id)
        ids = append(ids, id)
    }

    uniqueIDs := make(map[uuid.UUID]bool)
    for _, id := range ids {
        uniqueIDs[id] = true
    }
    assert.Equal(t, len(names), len(uniqueIDs))

    var count int64
    db.Model(&gormmodels.Room{}).Count(&count)
    assert.Equal(t, int64(len(names)), count)
}

func TestClassroomRepository_GetOrCreateByName_SpecialCharacters(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()

    repo := gormrepository.NewClassroomRepository(db)
    ctx := context.Background()

    specialName := "Room!@#$%^&*()_+{}[]|\\:;\"'<>,.?/~`"

    roomID, err := repo.GetOrCreateByName(ctx, specialName)
    assert.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, roomID)

    classroom, err := repo.GetByID(ctx, roomID)
    assert.NoError(t, err)
    assert.Equal(t, specialName, classroom.Name)
}