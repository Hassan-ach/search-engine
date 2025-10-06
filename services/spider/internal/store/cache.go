package store

import (
	"context"
	"log"

	"github.com/redis/go-redis/v9"
)

var (
	ctx    context.Context = context.Background()
	client                 = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol
	})
)

func NewCacheClient() *redis.Client {
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connected to redis %v", err)
	}
	return client
}

func NewHostName() (string, error) {
	s, err := client.SPop(ctx, "newHost").Result()
	if err != nil {
		return "", err
	}
	return s, nil
}
