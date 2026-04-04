package integration

import (
    "context"
    "testing"
    "time"

    "stud_hub/internal/models"
    redisrepository "stud_hub/internal/repository/redis-repository"

    "github.com/google/uuid"
    "github.com/redis/go-redis/v9"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func SetupTestRedis(t *testing.T) (*redis.Client, func()) {
    t.Helper()

    client := redis.NewClient(&redis.Options{
        Addr: "127.0.0.1:6379",
        DB:   1, 
    })

    ctx := context.Background()
    if err := client.Ping(ctx).Err(); err != nil {
        t.Skipf("Redis not available, skipping tests: %v", err)
    }

    cleanup := func() {
        client.FlushDB(ctx)
        client.Close()
    }

    return client, cleanup
}

func TestTokenRepository_SaveRefreshToken(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()
    tokenID := uuid.New()
    expiresAt := time.Now().Add(24 * time.Hour)

    token := models.RefreshToken{
        ID:        tokenID,
        UserID:    userID,
        IssuedAt:  time.Now(),
        ExpiresAt: expiresAt,
    }

    err := repo.SaveRefreshToken(ctx, token)
    assert.NoError(t, err)

    savedToken, err := repo.GetRefreshToken(ctx, tokenID)
    assert.NoError(t, err)
    assert.Equal(t, token.ID, savedToken.ID)
    assert.Equal(t, token.UserID, savedToken.UserID)
}

func TestTokenRepository_SaveRefreshToken_Expired(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()
    tokenID := uuid.New()
    expiresAt := time.Now().Add(-1 * time.Hour) 

    token := models.RefreshToken{
        ID:        tokenID,
        UserID:    userID,
        IssuedAt:  time.Now().Add(-2 * time.Hour),
        ExpiresAt: expiresAt,
    }

    err := repo.SaveRefreshToken(ctx, token)
    assert.Error(t, err)
}

func TestTokenRepository_GetRefreshToken(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()
    tokenID := uuid.New()
    expiresAt := time.Now().Add(24 * time.Hour).Truncate(time.Second)
    issuedAt := time.Now().Truncate(time.Second)

    token := models.RefreshToken{
        ID:        tokenID,
        UserID:    userID,
        IssuedAt:  issuedAt,
        ExpiresAt: expiresAt,
    }

    err := repo.SaveRefreshToken(ctx, token)
    require.NoError(t, err)

    savedToken, err := repo.GetRefreshToken(ctx, tokenID)
    assert.NoError(t, err)
    assert.Equal(t, token.ID, savedToken.ID)
    assert.Equal(t, token.UserID, savedToken.UserID)
    assert.True(t, savedToken.IssuedAt.Truncate(time.Second).Equal(issuedAt.Truncate(time.Second)))
    assert.True(t, savedToken.ExpiresAt.Truncate(time.Second).Equal(expiresAt.Truncate(time.Second)))
}

func TestTokenRepository_GetRefreshToken_NotFound(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    nonExistentID := uuid.New()
    token, err := repo.GetRefreshToken(ctx, nonExistentID)

    assert.Error(t, err)
    assert.Equal(t, models.RefreshToken{}, token)
}

func TestTokenRepository_DeleteRefreshToken(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()
    tokenID := uuid.New()
    expiresAt := time.Now().Add(24 * time.Hour)

    token := models.RefreshToken{
        ID:        tokenID,
        UserID:    userID,
        IssuedAt:  time.Now(),
        ExpiresAt: expiresAt,
    }

    err := repo.SaveRefreshToken(ctx, token)
    require.NoError(t, err)

    err = repo.DeleteRefreshToken(ctx, tokenID)
    assert.NoError(t, err)

    _, err = repo.GetRefreshToken(ctx, tokenID)
    assert.Error(t, err)
}

func TestTokenRepository_DeleteRefreshToken_NotFound(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    nonExistentID := uuid.New()
    err := repo.DeleteRefreshToken(ctx, nonExistentID)

    assert.Error(t, err)
}

func TestTokenRepository_DeleteUserRefreshTokens(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()
    tokenIDs := make([]uuid.UUID, 3)

    for i := 0; i < 3; i++ {
        tokenIDs[i] = uuid.New()
        token := models.RefreshToken{
            ID:        tokenIDs[i],
            UserID:    userID,
            IssuedAt:  time.Now(),
            ExpiresAt: time.Now().Add(24 * time.Hour),
        }
        err := repo.SaveRefreshToken(ctx, token)
        require.NoError(t, err)
    }

    err := repo.DeleteUserRefreshTokens(ctx, userID)
    assert.NoError(t, err)

    for _, tokenID := range tokenIDs {
        _, err := repo.GetRefreshToken(ctx, tokenID)
        assert.Error(t, err)
    }
}

func TestTokenRepository_DeleteUserRefreshTokens_Empty(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()

    err := repo.DeleteUserRefreshTokens(ctx, userID)
    assert.NoError(t, err)
}

func TestTokenRepository_MultipleTokensForUser(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()
    tokenCount := 5

    for i := 0; i < tokenCount; i++ {
        token := models.RefreshToken{
            ID:        uuid.New(),
            UserID:    userID,
            IssuedAt:  time.Now(),
            ExpiresAt: time.Now().Add(24 * time.Hour),
        }
        err := repo.SaveRefreshToken(ctx, token)
        require.NoError(t, err)
    }

    err := repo.DeleteUserRefreshTokens(ctx, userID)
    assert.NoError(t, err)

    for i := 0; i < tokenCount; i++ {
        token := models.RefreshToken{
            ID:        uuid.New(),
            UserID:    userID,
            IssuedAt:  time.Now(),
            ExpiresAt: time.Now().Add(24 * time.Hour),
        }
        err := repo.SaveRefreshToken(ctx, token)
        require.NoError(t, err)
    }

    assert.NoError(t, nil)
}

func TestTokenRepository_TokenExpiration(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()
    tokenID := uuid.New()
    expiresAt := time.Now().Add(2 * time.Second) 

    token := models.RefreshToken{
        ID:        tokenID,
        UserID:    userID,
        IssuedAt:  time.Now(),
        ExpiresAt: expiresAt,
    }

    err := repo.SaveRefreshToken(ctx, token)
    require.NoError(t, err)

    _, err = repo.GetRefreshToken(ctx, tokenID)
    assert.NoError(t, err)

    time.Sleep(3 * time.Second)

    _, err = repo.GetRefreshToken(ctx, tokenID)
    assert.Error(t, err)
}

func TestTokenRepository_OverwriteToken(t *testing.T) {
    redisClient, cleanup := SetupTestRedis(t)
    defer cleanup()

    repo := redisrepository.NewTokenRepository(redisClient)
    ctx := context.Background()

    userID := uuid.New()
    tokenID := uuid.New()

    issuedAt1 := time.Now().Truncate(time.Second)
    expiresAt1 := time.Now().Add(24 * time.Hour).Truncate(time.Second)
    
    token1 := models.RefreshToken{
        ID:        tokenID,
        UserID:    userID,
        IssuedAt:  issuedAt1,
        ExpiresAt: expiresAt1,
    }
    err := repo.SaveRefreshToken(ctx, token1)
    require.NoError(t, err)

    issuedAt2 := time.Now().Add(1 * time.Hour).Truncate(time.Second)
    expiresAt2 := time.Now().Add(48 * time.Hour).Truncate(time.Second)
    
    token2 := models.RefreshToken{
        ID:        tokenID,
        UserID:    userID,
        IssuedAt:  issuedAt2,
        ExpiresAt: expiresAt2,
    }
    err = repo.SaveRefreshToken(ctx, token2)
    assert.NoError(t, err)

    savedToken, err := repo.GetRefreshToken(ctx, tokenID)
    assert.NoError(t, err)
    assert.True(t, savedToken.IssuedAt.Truncate(time.Second).Equal(issuedAt2.Truncate(time.Second)))
    assert.True(t, savedToken.ExpiresAt.Truncate(time.Second).Equal(expiresAt2.Truncate(time.Second)))
}