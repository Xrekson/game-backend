package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/xrekson/game-backend/pkg/config"
	"github.com/xrekson/game-backend/pkg/logger"
)

type Client struct {
	*redis.Client
}

func NewRedisClient(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr(),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Warn("Redis ping failed, initial connection pending: %v", err)
	} else {
		logger.Info("Connected to Redis successfully at %s", cfg.Addr())
	}

	return &Client{Client: rdb}, nil
}
