package cache

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

var Redis *redis.Client
var Ctx = context.Background()

func InitRedis() {
	Redis = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_URL"), // e.g. localhost:6379
		Password: "",                     // or set from env
		DB:       0,
	})

	_, err := Redis.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
}

// CacheSet sets a key with TTL (seconds)
func CacheSet(key string, value string, ttl time.Duration) {
	Redis.Set(Ctx, key, value, ttl)
}

// CacheGet gets cached value if exists
func CacheGet(key string) (string, error) {
	return Redis.Get(Ctx, key).Result()
}
