package storage

import (
	"context"
	"fmt"

	"github.com/mildlybrutal/websocketGo/internal/common"
	"github.com/redis/go-redis/v9"
)

func InitRedis(cfg *common.RedisConfig) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		Protocol: 2,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}

	return rdb, nil
}
