package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RedisCache wraps redis client for caching operations
// Control: MOD-001 (Response caching for performance)
type RedisCache struct {
	client *redis.Client
	logger *zap.Logger
}

// Config holds Redis configuration
type Config struct {
	URL         string
	MaxRetries  int
	DialTimeout time.Duration
	ReadTimeout time.Duration
}

// NewRedisCache creates a new Redis client
func NewRedisCache(ctx context.Context, cfg Config, logger *zap.Logger) (*RedisCache, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	// Set configuration overrides
	if cfg.MaxRetries > 0 {
		opt.MaxRetries = cfg.MaxRetries
	} else {
		opt.MaxRetries = 3
	}

	if cfg.DialTimeout > 0 {
		opt.DialTimeout = cfg.DialTimeout
	} else {
		opt.DialTimeout = 5 * time.Second
	}

	if cfg.ReadTimeout > 0 {
		opt.ReadTimeout = cfg.ReadTimeout
	} else {
		opt.ReadTimeout = 3 * time.Second
	}

	client := redis.NewClient(opt)

	// Verify connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	logger.Info("redis connection established")

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

// Close closes the Redis client connection
func (r *RedisCache) Close() error {
	if r.client != nil {
		if err := r.client.Close(); err != nil {
			return fmt.Errorf("failed to close Redis client: %w", err)
		}
		r.logger.Info("redis connection closed")
	}
	return nil
}

// Health checks the Redis health
func (r *RedisCache) Health(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := r.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis health check failed: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", fmt.Errorf("key not found: %s", key)
	} else if err != nil {
		return "", fmt.Errorf("failed to get key %s: %w", key, err)
	}
	return val, nil
}

// Set stores a value in cache with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if err := r.client.Set(ctx, key, value, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set key %s: %w", key, err)
	}
	return nil
}

// Delete removes a key from cache
func (r *RedisCache) Delete(ctx context.Context, keys ...string) error {
	if err := r.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("failed to delete keys: %w", err)
	}
	return nil
}

// Exists checks if a key exists in cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to check key existence: %w", err)
	}
	return count > 0, nil
}

// Incr increments a counter
func (r *RedisCache) Incr(ctx context.Context, key string) (int64, error) {
	val, err := r.client.Incr(ctx, key).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to increment key %s: %w", key, err)
	}
	return val, nil
}

// Expire sets expiration on a key
func (r *RedisCache) Expire(ctx context.Context, key string, expiration time.Duration) error {
	if err := r.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("failed to set expiration on key %s: %w", key, err)
	}
	return nil
}
