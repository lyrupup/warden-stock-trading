// Package cache 封装 Redis 客户端的初始化与连接探测。
package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// Config 是 Redis 连接配置。
type Config struct {
	Host     string
	Port     int
	Password string
	DB       int
}

// New 创建 Redis 客户端并做一次 Ping 探测。失败返回 error，
// 由调用方决定降级策略，不在此处 panic。
func New(cfg Config) (*redis.Client, error) {
	cli := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		return cli, err
	}
	return cli, nil
}
