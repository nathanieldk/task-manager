package redis

import (
	"context"
	"fmt"

	goredis "github.com/redis/go-redis/v9"
)

// Connect establishes a connection to Redis and pings it.
func Connect(ctx context.Context, addr string, password string, db int) (*goredis.Client, error) {
	rdb := goredis.NewClient(&goredis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		rdb.Close()
		return nil, fmt.Errorf("failed to ping redis: %w", err)
	}

	return rdb, nil
}
