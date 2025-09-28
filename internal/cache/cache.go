package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// Cache interface for caching operations
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Delete(ctx context.Context, key string) error
	Close() error
}

// RedisCache implements Cache using Redis
type RedisCache struct {
	client *redis.Client
	logger *logrus.Logger
}

// NewRedisCache creates a new Redis cache instance
func NewRedisCache(url, password string, db int, logger *logrus.Logger) (Cache, error) {
	opts := &redis.Options{
		Addr:     "localhost:6379", // Extract from URL in real implementation
		Password: password,
		DB:       db,
	}

	client := redis.NewClient(opts)

	// Test connection
	ctx := context.Background()
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisCache{
		client: client,
		logger: logger,
	}, nil
}

func (r *RedisCache) Get(ctx context.Context, key string) (string, error) {
	result := r.client.Get(ctx, key)
	if err := result.Err(); err != nil {
		if err == redis.Nil {
			return "", nil // Key not found
		}
		return "", err
	}
	return result.Val(), nil
}

func (r *RedisCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		str = string(data)
	}

	return r.client.Set(ctx, key, str, expiration).Err()
}

func (r *RedisCache) Delete(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisCache) Close() error {
	return r.client.Close()
}

// MemoryCache implements Cache using in-memory storage
type MemoryCache struct {
	data   map[string]string
	expiry map[string]time.Time
}

// NewMemoryCache creates a new in-memory cache
func NewMemoryCache() Cache {
	return &MemoryCache{
		data:   make(map[string]string),
		expiry: make(map[string]time.Time),
	}
}

func (m *MemoryCache) Get(ctx context.Context, key string) (string, error) {
	if expiry, exists := m.expiry[key]; exists {
		if time.Now().After(expiry) {
			delete(m.data, key)
			delete(m.expiry, key)
			return "", nil
		}
	}
	return m.data[key], nil
}

func (m *MemoryCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	var str string
	switch v := value.(type) {
	case string:
		str = v
	default:
		data, err := json.Marshal(value)
		if err != nil {
			return err
		}
		str = string(data)
	}

	m.data[key] = str
	if expiration > 0 {
		m.expiry[key] = time.Now().Add(expiration)
	}
	return nil
}

func (m *MemoryCache) Delete(ctx context.Context, key string) error {
	delete(m.data, key)
	delete(m.expiry, key)
	return nil
}

func (m *MemoryCache) Close() error {
	return nil
}
