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

func TestUserRepository_CreateUser(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()
    
    err := ClearTables(db, "users")
    require.NoError(t, err)
    
    repo := gormrepository.NewUserRepository(db)
    ctx := context.Background()
    
    user := models.User{
        ID:           uuid.New(),
        Email:        "test@example.com",
        PasswordHash: "hashed_password_123", 
        FirstName:    "Test",
        LastName:     "User",
        Patronymic:   "Testovich",
        Role:         models.RoleStudent,
        IsBlocked:    false,
        GroupID:      uuid.Nil, 
        UniversityID: uuid.Nil,
    }
    
    userID, err := repo.CreateUser(ctx, user)
    
    assert.NoError(t, err)
    assert.Equal(t, user.ID, userID)
    
    var savedUser gormmodels.User
    err = db.WithContext(ctx).First(&savedUser, "id = ?", userID).Error
    assert.NoError(t, err)
    assert.Equal(t, user.Email, savedUser.Email)
    assert.Equal(t, user.FirstName, savedUser.FirstName)
    assert.Equal(t, user.LastName, savedUser.LastName)
    assert.Equal(t, user.Patronymic, savedUser.Patronymic)
    assert.NotNil(t, savedUser.PasswordHash)
    assert.Equal(t, user.PasswordHash, *savedUser.PasswordHash)
}

func TestUserRepository_GetUserByID(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()
    
    err := ClearTables(db, "users")
    require.NoError(t, err)
    
    passwordHash := "hashed_password"
    
    testUser := gormmodels.User{
        ID:           uuid.New(),
        Email:        "gettest@example.com",
        PasswordHash: &passwordHash,
        FirstName:    "Get",
        LastName:     "Test",
        Patronymic:   "Getovich",
        Role:         gormmodels.UserRoleStudent,
        IsBlocked:    false,
        GroupID:      nil,
    }
    err = db.Create(&testUser).Error
    require.NoError(t, err)
    
    repo := gormrepository.NewUserRepository(db)
    ctx := context.Background()
    
    user, err := repo.GetUserByID(ctx, testUser.ID)
    
    assert.NoError(t, err)
    assert.Equal(t, testUser.ID, user.ID)
    assert.Equal(t, testUser.Email, user.Email)
    assert.Equal(t, testUser.FirstName, user.FirstName)
    assert.Equal(t, testUser.LastName, user.LastName)
    assert.Equal(t, testUser.Patronymic, user.Patronymic)
    assert.Equal(t, passwordHash, user.PasswordHash)
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()
    
    err := ClearTables(db, "users")
    require.NoError(t, err)
    
    repo := gormrepository.NewUserRepository(db)
    ctx := context.Background()
    
    nonExistentID := uuid.New()
    user, err := repo.GetUserByID(ctx, nonExistentID)
    
    assert.Error(t, err)
    assert.Equal(t, models.User{}, user)
}

func TestUserRepository_UpdateUser(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()
    
    err := ClearTables(db, "users")
    require.NoError(t, err)
    
    oldPasswordHash := "old_hashed_password"
    
    existingUser := gormmodels.User{
        ID:           uuid.New(),
        Email:        "update@example.com",
        PasswordHash: &oldPasswordHash,
        FirstName:    "Old",
        LastName:     "Name",
        Patronymic:   "Oldovich",
        Role:         gormmodels.UserRoleStudent,
        IsBlocked:    false,
        GroupID:      nil,
    }
    err = db.Create(&existingUser).Error
    require.NoError(t, err)
    
    repo := gormrepository.NewUserRepository(db)
    ctx := context.Background()
    
    newPasswordHash := "new_hashed_password"
    updatedUser := models.User{
        ID:           existingUser.ID,
        Email:        "updated@example.com",
        PasswordHash: newPasswordHash,
        FirstName:    "New",
        LastName:     "Name",
        Patronymic:   "Newovich",
        Role:         models.RoleAdmin,
        IsBlocked:    true,
        GroupID:      uuid.Nil,
        UniversityID: uuid.Nil,
    }
    
    result, err := repo.UpdateUser(ctx, updatedUser)
    
    assert.NoError(t, err)
    assert.Equal(t, updatedUser.Email, result.Email)
    assert.Equal(t, updatedUser.FirstName, result.FirstName)
    assert.Equal(t, updatedUser.LastName, result.LastName)
    assert.Equal(t, updatedUser.Patronymic, result.Patronymic)
    assert.Equal(t, updatedUser.Role, result.Role)
    assert.Equal(t, updatedUser.IsBlocked, result.IsBlocked)
    
    var dbUser gormmodels.User
    err = db.First(&dbUser, "id = ?", existingUser.ID).Error
    assert.NoError(t, err)
    assert.Equal(t, "updated@example.com", dbUser.Email)
    assert.Equal(t, "New", dbUser.FirstName)
    assert.Equal(t, true, dbUser.IsBlocked)
    assert.NotNil(t, dbUser.PasswordHash)
    assert.Equal(t, newPasswordHash, *dbUser.PasswordHash)
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()
    
    err := ClearTables(db, "users")
    require.NoError(t, err)
    
    passwordHash := "hashed_password"
    
    testUser := gormmodels.User{
        ID:           uuid.New(),
        Email:        "emailtest@example.com",
        PasswordHash: &passwordHash,
        FirstName:    "Email",
        LastName:     "Test",
        Patronymic:   "Emailovich",
        Role:         gormmodels.UserRoleStudent,
        IsBlocked:    false,
        GroupID:      nil,
    }
    err = db.Create(&testUser).Error
    require.NoError(t, err)
    
    repo := gormrepository.NewUserRepository(db)
    ctx := context.Background()
    
    user, err := repo.GetUserByEmail(ctx, "emailtest@example.com")
    
    assert.NoError(t, err)
    assert.Equal(t, testUser.ID, user.ID)
    assert.Equal(t, testUser.Email, user.Email)
    assert.Equal(t, testUser.FirstName, user.FirstName)
    assert.Equal(t, testUser.LastName, user.LastName)
}

func TestUserRepository_CreateUser_DuplicateEmail(t *testing.T) {
    db, cleanup := SetupTestDB(t)
    defer cleanup()
    
    err := ClearTables(db, "users")
    require.NoError(t, err)
    
    repo := gormrepository.NewUserRepository(db)
    ctx := context.Background()
    
    passwordHash := "hashed_password"
    
    user1 := models.User{
        ID:           uuid.New(),
        Email:        "duplicate@example.com",
        PasswordHash: passwordHash,
        FirstName:    "First",
        LastName:     "User",
        Role:         models.RoleStudent,
        IsBlocked:    false,
    }
    
    _, err = repo.CreateUser(ctx, user1)
    require.NoError(t, err)
    
    user2 := models.User{
        ID:           uuid.New(),
        Email:        "duplicate@example.com",
        PasswordHash: passwordHash,
        FirstName:    "Second",
        LastName:     "User",
        Role:         models.RoleStudent,
        IsBlocked:    false,
    }
    
    _, err = repo.CreateUser(ctx, user2)
    
    assert.Error(t, err)
}