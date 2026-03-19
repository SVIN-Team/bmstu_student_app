package redisrepository

import (
	"context"
	"encoding/json"
	"time"

	"stud_hub/internal/config"
	autherrors "stud_hub/internal/errors"
	"stud_hub/internal/models"
	"stud_hub/util/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	refreshTokenKeyPrefix = "refresh_token:"
	userTokensKeyPrefix   = "user_refresh_tokens:"
)

func InitRedis(ctx context.Context, config *config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
			Addr:     config.RedisServer,    // Redis server address
			Password: config.RedisPassword,  // No password set
			DB:       config.DatabaseNumber, // Use the default DB
		})
	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Errorf(ctx, "could not connect to redis: %v", err)
		return nil, err
	}

	logger.Infof(ctx, "successfully connected to redis")
	return client, nil
}

// TokenRepository stores refresh tokens in Redis.
type TokenRepository struct {
	client *redis.Client
}

func NewTokenRepository(client *redis.Client) *TokenRepository {
	return &TokenRepository{client: client}
}

type refreshTokenPayload struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

func refreshTokenKey(jti uuid.UUID) string {
	return refreshTokenKeyPrefix + jti.String()
}

func userTokensKey(userID uuid.UUID) string {
	return userTokensKeyPrefix + userID.String()
}

func (r *TokenRepository) SaveRefreshToken(ctx context.Context, token models.RefreshToken) error {
	ttl := time.Until(token.ExpiresAt)
	if ttl <= 0 {
		logger.Errorf(ctx, "attempted to save expired refresh token %s", token.ID)
		return autherrors.ErrInternalServer
	}

	payload := refreshTokenPayload{
		ID:        token.ID.String(),
		UserID:    token.UserID.String(),
		IssuedAt:  token.IssuedAt,
		ExpiresAt: token.ExpiresAt,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		logger.Errorf(ctx, "failed to marshal refresh token %s: %v", token.ID, err)
		return autherrors.ErrInternalServer
	}

	pipe := r.client.TxPipeline()
	pipe.Set(ctx, refreshTokenKey(token.ID), data, ttl)
	pipe.SAdd(ctx, userTokensKey(token.UserID), token.ID.String())
	pipe.Expire(ctx, userTokensKey(token.UserID), ttl)

	if _, err = pipe.Exec(ctx); err != nil {
		logger.Errorf(ctx, "failed to save refresh token %s in redis: %v", token.ID, err)
		return autherrors.ErrInternalServer
	}
	return nil
}

func (r *TokenRepository) GetRefreshToken(ctx context.Context, jti uuid.UUID) (models.RefreshToken, error) {
	val, err := r.client.Get(ctx, refreshTokenKey(jti)).Result()
	if err == redis.Nil {
		return models.RefreshToken{}, autherrors.ErrInvalidRefreshToken
	}
	if err != nil {
		logger.Errorf(ctx, "failed to get refresh token %s: %v", jti, err)
		return models.RefreshToken{}, autherrors.ErrInternalServer
	}

	var payload refreshTokenPayload
	if err := json.Unmarshal([]byte(val), &payload); err != nil {
		logger.Errorf(ctx, "failed to unmarshal refresh token %s: %v", jti, err)
		return models.RefreshToken{}, autherrors.ErrInternalServer
	}

	userID, err := uuid.Parse(payload.UserID)
	if err != nil {
		logger.Errorf(ctx, "failed to parse user id %s for token %s: %v", payload.UserID, jti, err)
		return models.RefreshToken{}, autherrors.ErrInternalServer
	}

	return models.RefreshToken{
		ID:        jti,
		UserID:    userID,
		IssuedAt:  payload.IssuedAt,
		ExpiresAt: payload.ExpiresAt,
	}, nil
}

func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, jti uuid.UUID) error {
	token, err := r.GetRefreshToken(ctx, jti)
	if err != nil {
		return err
	}

	pipe := r.client.TxPipeline()
	pipe.Del(ctx, refreshTokenKey(jti))
	pipe.SRem(ctx, userTokensKey(token.UserID), jti.String())

	if _, err = pipe.Exec(ctx); err != nil {
		logger.Errorf(ctx, "failed to delete refresh token %s: %v", jti, err)
		return autherrors.ErrInternalServer
	}
	return nil
}

func (r *TokenRepository) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	jtis, err := r.client.SMembers(ctx, userTokensKey(userID)).Result()
	if err != nil && err != redis.Nil {
		logger.Errorf(ctx, "failed to list refresh tokens for user %s: %v", userID, err)
		return autherrors.ErrInternalServer
	}

	pipe := r.client.TxPipeline()
	for _, jti := range jtis {
		pipe.Del(ctx, refreshTokenKeyPrefix+jti)
	}
	pipe.Del(ctx, userTokensKey(userID))

	if _, err = pipe.Exec(ctx); err != nil {
		logger.Errorf(ctx, "failed to delete refresh tokens for user %s: %v", userID, err)
		return autherrors.ErrInternalServer
	}
	return nil
}
