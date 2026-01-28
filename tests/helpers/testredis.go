package helpers

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestRedis provides a test Redis setup
type TestRedis struct {
	Client *redis.Client
	Addr   string
}

// SetupTestRedis creates a test Redis connection
func SetupTestRedis(t *testing.T) (*TestRedis, func()) {
	t.Helper()

	// Get Redis address from environment or use default
	addr := os.Getenv("TEST_REDIS_ADDR")
	if addr == "" {
		addr = "localhost:6379"
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     "",
		DB:           1, // Use DB 1 for tests to avoid conflicts
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 2,
	})

	ctx := context.Background()

	// Test connection
	if err := client.Ping(ctx).Err(); err != nil {
		client.Close()
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	testRedis := &TestRedis{
		Client: client,
		Addr:   addr,
	}

	// Flush test DB to ensure clean state
	if err := client.FlushDB(ctx).Err(); err != nil {
		client.Close()
		t.Fatalf("Failed to flush Redis DB: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		testRedis.Cleanup(t)
		client.Close()
	}

	return testRedis, cleanup
}

// Cleanup flushes all data from the test Redis database
func (r *TestRedis) Cleanup(t *testing.T) {
	t.Helper()

	ctx := context.Background()
	if err := r.Client.FlushDB(ctx).Err(); err != nil {
		t.Logf("Warning: Failed to flush Redis DB: %v", err)
	}
}

// SetTestCache sets a test cache value
func (r *TestRedis) SetTestCache(t *testing.T, key string, value interface{}, expiration time.Duration) {
	t.Helper()

	ctx := context.Background()
	if err := r.Client.Set(ctx, key, value, expiration).Err(); err != nil {
		t.Fatalf("Failed to set test cache: %v", err)
	}
}

// GetTestCache retrieves a test cache value
func (r *TestRedis) GetTestCache(t *testing.T, key string) string {
	t.Helper()

	ctx := context.Background()
	val, err := r.Client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return ""
		}
		t.Fatalf("Failed to get test cache: %v", err)
	}
	return val
}

// KeyExists checks if a key exists in Redis
func (r *TestRedis) KeyExists(t *testing.T, key string) bool {
	t.Helper()

	ctx := context.Background()
	exists, err := r.Client.Exists(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to check key existence: %v", err)
	}
	return exists > 0
}

// GetTTL returns the TTL of a key
func (r *TestRedis) GetTTL(t *testing.T, key string) time.Duration {
	t.Helper()

	ctx := context.Background()
	ttl, err := r.Client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("Failed to get TTL: %v", err)
	}
	return ttl
}

// CountKeys returns the number of keys matching a pattern
func (r *TestRedis) CountKeys(t *testing.T, pattern string) int {
	t.Helper()

	ctx := context.Background()
	keys, err := r.Client.Keys(ctx, pattern).Result()
	if err != nil {
		t.Fatalf("Failed to count keys: %v", err)
	}
	return len(keys)
}

// DeleteKeys deletes all keys matching a pattern
func (r *TestRedis) DeleteKeys(t *testing.T, pattern string) {
	t.Helper()

	ctx := context.Background()
	keys, err := r.Client.Keys(ctx, pattern).Result()
	if err != nil {
		t.Fatalf("Failed to get keys: %v", err)
	}

	if len(keys) > 0 {
		if err := r.Client.Del(ctx, keys...).Err(); err != nil {
			t.Fatalf("Failed to delete keys: %v", err)
		}
	}
}
