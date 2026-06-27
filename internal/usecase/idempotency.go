package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nathanieldk/task-manager/internal/entity"

	"github.com/redis/go-redis/v9"
)

const (
	idempotencyPrefix = "idempotency:"
	idempotencyTTL    = 24 * time.Hour
	lockSuffix        = ":lock"
	lockTTL           = 30 * time.Second
)

type IdempotencyUsecase interface {
	// Acquire attempts to acquire a lock for the given idempotency key.
	// Returns true if the lock was acquired (this is the first request).
	Acquire(ctx context.Context, key string) (bool, error)

	// Get retrieves a cached task response for the given idempotency key.
	// Returns nil if no cached response exists.
	Get(ctx context.Context, key string) (*entity.Task, error)

	// Store caches a task response for the given idempotency key with a 24h TTL.
	Store(ctx context.Context, key string, task *entity.Task) error
}

type redisIdempotencyUsecase struct {
	client *redis.Client
}

// NewIdempotencyUsecase creates a new Redis-backed idempotency usecase.
func NewIdempotencyUsecase(client *redis.Client) IdempotencyUsecase {
	return &redisIdempotencyUsecase{client: client}
}

func (s *redisIdempotencyUsecase) Acquire(ctx context.Context, key string) (bool, error) {
	lockKey := fmt.Sprintf("%s%s%s", idempotencyPrefix, key, lockSuffix)

	// SETNX — atomic set-if-not-exists
	result, err := s.client.SetNX(ctx, lockKey, "locked", lockTTL).Result()
	if err != nil {
		return false, fmt.Errorf("failed to acquire idempotency lock: %w", err)
	}

	return result, nil
}

func (s *redisIdempotencyUsecase) Get(ctx context.Context, key string) (*entity.Task, error) {
	dataKey := fmt.Sprintf("%s%s", idempotencyPrefix, key)

	data, err := s.client.Get(ctx, dataKey).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get idempotency data: %w", err)
	}

	var task entity.Task
	if err := json.Unmarshal(data, &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal idempotency data: %w", err)
	}

	return &task, nil
}

func (s *redisIdempotencyUsecase) Store(ctx context.Context, key string, task *entity.Task) error {
	dataKey := fmt.Sprintf("%s%s", idempotencyPrefix, key)

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task for idempotency: %w", err)
	}

	if err := s.client.Set(ctx, dataKey, data, idempotencyTTL).Err(); err != nil {
		return fmt.Errorf("failed to store idempotency data: %w", err)
	}

	return nil
}
