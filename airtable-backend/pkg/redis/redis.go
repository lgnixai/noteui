package redis

import (
	"context"
	"log"
	"time"

	"airtable-backend/configs"

	"github.com/go-redis/redis/v8"
)

var RDB *redis.Client
var Ctx = context.Background()

func ConnectRedis(cfg *configs.Config) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		log.Fatalf("Failed to parse Redis URL: %v", err)
	}

	RDB = redis.NewClient(opt)

	// Ping to check connection
	ctx, cancel := context.WithTimeout(Ctx, 5*time.Second)
	defer cancel()
	_, err = RDB.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
}
