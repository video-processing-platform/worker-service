package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alimarzban99/video-processor-service/config"
	"github.com/redis/go-redis/v9"
)

var (
	client *redis.Client
	once   sync.Once
)

func Init() error {
	var initErr error

	once.Do(func() {
		cfg := config.Cfg.Redis

		client = redis.NewClient(&redis.Options{
			Addr:         fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
			Password:     cfg.Password,
			DB:           cfg.Database,
			DialTimeout:  5 * time.Second,
			ReadTimeout:  5 * time.Second,
			WriteTimeout: 5 * time.Second,
			PoolSize:     cfg.PoolSize,
			PoolTimeout:  15 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := client.Ping(ctx).Err(); err != nil {
			initErr = err
			return
		}

		log.Printf("✅ Redis connected (%s:%s)", cfg.Host, cfg.Port)
	})

	return initErr
}

// Client returns initialized redis client
func Client() (*redis.Client, error) {
	if client == nil {
		return nil, errors.New("redis not initialized")
	}
	return client, nil
}

// Close closes redis connection
func Close() error {
	if client == nil {
		return nil
	}

	log.Println("🛑 Redis connection closed")
	return client.Close()
}

/*
|--------------------------------------------------------------------------
| Helpers
|--------------------------------------------------------------------------
*/

func Set[T any](ctx context.Context, key string, value T, ttl time.Duration) error {
	if client == nil {
		return errors.New("redis not initialized")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return client.Set(ctx, key, data, ttl).Err()
}

func Get[T any](ctx context.Context, key string) (T, error) {
	var dest T

	if client == nil {
		return dest, errors.New("redis not initialized")
	}

	val, err := client.Get(ctx, key).Result()
	if err != nil {
		return dest, err
	}

	if err := json.Unmarshal([]byte(val), &dest); err != nil {
		return dest, err
	}

	return dest, nil
}
