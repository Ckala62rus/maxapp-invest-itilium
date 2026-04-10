package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache wraps go-redis and hides cache operations behind small helpers.
type RedisCache struct {
	client  *redis.Client
	enabled bool
}

// NewRedisCache creates a cache wrapper and keeps it disabled when no client exists.
func NewRedisCache(client *redis.Client) *RedisCache {
	return &RedisCache{
		client:  client,
		enabled: client != nil,
	}
}

// GetJSON reads a JSON document from Redis into the provided target.
func (c *RedisCache) GetJSON(ctx context.Context, key string, target any) (bool, error) {
	if !c.enabled {
		return false, nil
	}

	value, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal([]byte(value), target); err != nil {
		return false, err
	}

	return true, nil
}

// SetJSON stores a JSON document in Redis with a TTL.
func (c *RedisCache) SetJSON(ctx context.Context, key string, value any, ttl time.Duration) error {
	if !c.enabled {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return c.client.Set(ctx, key, data, ttl).Err()
}
