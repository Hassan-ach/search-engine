package store

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

type CacheClient struct{ *redis.Client }

var (
	ctx   context.Context = context.Background()
	Cache                 = NewCacheClient()
)

func NewCacheClient() *CacheClient {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})

	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connected to redis %v", err)
	}
	return &CacheClient{client}
}

func PopHost() (string, error) {
	s, err := Cache.SPop(ctx, "newHost").Result()
	if err != nil {
		return "", err
	}
	return s, nil
}
