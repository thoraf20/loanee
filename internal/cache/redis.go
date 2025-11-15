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
		Addr:     os.Getenv("REDIS_URL"),
		Password: "",                   
		DB:       0,
	})

	_, err := Redis.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
}

func CacheSet(key, value string, ttl time.Duration) {
	Redis.Set(Ctx, key, value, ttl)
}

func CacheGet(key string) (string, error) {
	return Redis.Get(Ctx, key).Result()
}

func CacheDelete(key string) error {
	return Redis.Del(Ctx, key).Err()
}