package services

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type CacheService interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, key string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Ping(ctx context.Context) error
	IsEnabled() bool
}

type RedisCache struct {
	client  *redis.Client
	enabled bool
}

// NewRedisCache creates a new Redis cache service with connection testing
func NewRedisCache(addr, password string, db int) *RedisCache {
	client := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     password,
		DB:           db,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
		PoolSize:     10,
		MinIdleConns: 5,
	})

	cache := &RedisCache{
		client:  client,
		enabled: false,
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Redis cache unavailable: %v - Caching disabled", err)
		return cache
	}

	cache.enabled = true
	log.Println("✓ Redis cache connected successfully")
	return cache
}

// IsEnabled returns whether the cache is available
func (r *RedisCache) IsEnabled() bool {
	return r.enabled
}

// Set stores a value in cache with expiration
func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if !r.enabled {
		return nil // Silently skip if cache is disabled
	}

	data, err := json.Marshal(value)
	if err != nil {
		log.Printf("Cache marshal error for key %s: %v", key, err)
		return fmt.Errorf("failed to marshal cache value: %w", err)
	}

	if err := r.client.Set(ctx, key, data, expiration).Err(); err != nil {
		log.Printf("Cache set error for key %s: %v", key, err)
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// Get retrieves a value from cache
func (r *RedisCache) Get(ctx context.Context, key string, dest interface{}) error {
	if !r.enabled {
		return redis.Nil // Return cache miss if disabled
	}

	data, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			// Cache miss is not an error - it's expected behavior
			return redis.Nil
		}
		log.Printf("Cache get error for key %s: %v", key, err)
		return fmt.Errorf("failed to get cache: %w", err)
	}

	if err := json.Unmarshal([]byte(data), dest); err != nil {
		log.Printf("Cache unmarshal error for key %s: %v", key, err)
		return fmt.Errorf("failed to unmarshal cache value: %w", err)
	}

	return nil
}

// Delete removes a key from cache
func (r *RedisCache) Delete(ctx context.Context, key string) error {
	if !r.enabled {
		return nil
	}

	if err := r.client.Del(ctx, key).Err(); err != nil {
		log.Printf("Cache delete error for key %s: %v", key, err)
		return fmt.Errorf("failed to delete cache: %w", err)
	}

	return nil
}

// DeletePattern deletes all keys matching a pattern (e.g., "user:*")
func (r *RedisCache) DeletePattern(ctx context.Context, pattern string) error {
	if !r.enabled {
		return nil
	}

	iter := r.client.Scan(ctx, 0, pattern, 100).Iterator()
	keys := []string{}

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
	}

	if err := iter.Err(); err != nil {
		log.Printf("Cache scan error for pattern %s: %v", pattern, err)
		return fmt.Errorf("failed to scan cache keys: %w", err)
	}

	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			log.Printf("Cache delete pattern error for %s: %v", pattern, err)
			return fmt.Errorf("failed to delete cache pattern: %w", err)
		}
		log.Printf("Invalidated %d cache keys matching pattern: %s", len(keys), pattern)
	}

	return nil
}

// Exists checks if a key exists in cache
func (r *RedisCache) Exists(ctx context.Context, key string) (bool, error) {
	if !r.enabled {
		return false, nil
	}

	count, err := r.client.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// Ping tests the cache connection
func (r *RedisCache) Ping(ctx context.Context) error {
	if !r.enabled {
		return fmt.Errorf("cache is disabled")
	}

	return r.client.Ping(ctx).Err()
}
