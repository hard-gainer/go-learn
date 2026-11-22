package cache

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type Redis struct {
	client *redis.Client
}

func Open(ctx context.Context) (*Redis, error) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}
	var db int
	if value := os.Getenv("REDIS_DB"); value != "" {
		index, err := strconv.Atoi(value)
		if err != nil || index < 0 {
			return nil, errors.New("REDIS_DB must be a non-negative integer")
		}
		db = index
	}

	client := redis.NewClient(&redis.Options{
		Addr:                  addr,
		Password:              os.Getenv("REDIS_PASSWORD"),
		DB:                    db,
		DialTimeout:           time.Second,
		ReadTimeout:           time.Second,
		WriteTimeout:          time.Second,
		ContextTimeoutEnabled: true,
		MaxRetries:            -1,
	})
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("ping Redis at %s: %w", addr, err)
	}
	return &Redis{client: client}, nil
}

func (r *Redis) Close() error {
	return r.client.Close()
}

func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := r.client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return data, err
}

func (r *Redis) Set(ctx context.Context, key string, data []byte, ttl time.Duration) error {
	return r.client.Set(ctx, key, data, ttl).Err()
}
