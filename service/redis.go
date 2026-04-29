package service

import (
	"context"

	"github.com/jayson2hu/image-show/config"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

func RedisClient() *redis.Client {
	cfg := config.AppConfig
	if cfg == nil {
		cfg = config.LoadConfig()
	}
	if cfg.RedisAddr == "" {
		return nil
	}
	if redisClient == nil {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.RedisAddr,
			Password: cfg.RedisPassword,
			DB:       cfg.RedisDB,
		})
	}
	return redisClient
}

func CloseRedis() error {
	if redisClient == nil {
		return nil
	}
	err := redisClient.Close()
	redisClient = nil
	return err
}

func PingRedis(ctx context.Context) error {
	client := RedisClient()
	if client == nil {
		return nil
	}
	return client.Ping(ctx).Err()
}
