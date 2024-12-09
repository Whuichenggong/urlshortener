package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Whuichenggong/urlshortener/urlshortener/config"
	"github.com/Whuichenggong/urlshortener/urlshortener/internal/repo"
	"github.com/go-redis/redis/v8"
)

type RedisCache struct {
	client *redis.Client
}

func NewRedisCache(cfg config.RedisConfig) (*RedisCache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, err
	}
	return &RedisCache{client: client}, nil
}

func (c *RedisCache) SetURL(ctx context.Context, url repo.Url) error {
	data, err := json.Marshal(url)
	if err != nil {
		return fmt.Errorf("failed to marshal URL: %v", err)
	}
	if err := c.client.Set(ctx, url.ShortCode, data, time.Until(url.ExpiredAt)).Err(); err != nil {
		return err

	}
	return nil
}

func (c *RedisCache) GetURL(ctx context.Context, code string) (*repo.Url, error) {

	data, err := c.client.Get(ctx, code).Bytes()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var url repo.Url
	//反序列化
	if err := json.Unmarshal(data, &url); err != nil {
		return nil, err
	}
	return &url, nil
}

// 在 internal/cache/redis.go 中添加 Close 方法
func (c *RedisCache) Close() error {
	return c.client.Close()
}
