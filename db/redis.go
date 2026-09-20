package db

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client

func InitRedis(ctx context.Context) error {
	redisDB := 0

	if value := os.Getenv("REDIS_DB"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 0 {
			return fmt.Errorf("invalid REDIS_DB")
		}
		redisDB = parsed
	}

	address := os.Getenv("REDIS_ADDR")
	if address == "" {
		address = "localhost:6379"
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:     address,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       redisDB,
	})

	if err := Redis.Ping(ctx).Err(); err != nil {
		_ = Redis.Close()
		Redis = nil
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	return nil
}

func CloseRedis() error {
	if Redis == nil {
		return nil
	}

	return Redis.Close()
}
