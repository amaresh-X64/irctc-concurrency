package redis

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var Client *redis.Client
var Ctx = context.Background()

// ─── Connect to Redis ──────────────────────────
func Connect(redisURL string) *redis.Client {
	Client = redis.NewClient(&redis.Options{
		Addr: redisURL,
		DB:   0,
	})

	_, err := Client.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}

	log.Println("✅ Connected to Redis")
	return Client
}
