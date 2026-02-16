package store

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"spider/internal/config"
	"spider/internal/entity"
	"spider/internal/utils"
)

// NewCacheClient initializes and returns a Redis client.
// Registers entity.Host type with gob for serialization.
func NewCacheClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     config.RdAddr,
		Password: config.RdPassword,
		DB:       config.RdDb,
		Protocol: 2,
	})

	// Ping to ensure Redis connection is alive
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connected to redis ERROR: %v", err)
	}

	// Required for gob encoding/decoding of Host structs
	gob.Register(entity.Host{})
	fmt.Println("Cache Connected")
	return client
}

// AddHostMetaData serializes a Host struct with gob and stores it in Redis.
// Logs warnings for invalid input and errors for encoding/storage failures.
func AddHostMetaData(h string, host *entity.Host) error {
	if h == "" || host == nil {
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(host); err != nil {
		utils.Log.Cache().Debug("encode host metadata", "host", h, "error", err)
		return fmt.Errorf("encode metadata: %w", err)
	}

	err := Cache.HSet(ctx, "hosts", h, buf.Bytes()).Err()
	if err != nil {
		utils.Log.Cache().Debug("store host metadata", "host", h, "error", err)
		return fmt.Errorf("store metadata: %w", err)
	}

	return nil
}

// GetHostMetaData retrieves and decodes a Host struct from Redis.
// Logs each step: retrieval, missing host, decode error, successful fetch.
func GetHostMetaData(h string) (*entity.Host, bool, error) {
	val, err := Cache.HGet(ctx, "hosts", h).Bytes()
	if err == redis.Nil {
		utils.Log.Cache().Debug("host metadata not found", "host", h)
		return nil, false, fmt.Errorf("host metadata not found for host: %s", h)
	}

	if err != nil {
		utils.Log.Cache().Debug("retrieve host metadata", "host", h, "error", err)
		return nil, false, fmt.Errorf("retrieve host metadata: %w", err)
	}

	var host entity.Host
	if err := gob.NewDecoder(bytes.NewReader(val)).Decode(&host); err != nil {
		utils.Log.Cache().Debug("decode host metadata", "host", h, "error", err)
		return nil, false, fmt.Errorf("decode host metadata: %w", err)
	}

	return &host, true, nil
}

// GetUrl retrieves a URL from Redis atomically.
// Skips URLs if host is recently visited or URL is already visited.
// Uses Lua script to ensure atomicity.
func GetUrl() (string, bool, error) {
	script := redis.NewScript(`
	local res = redis.call("zpopmax", KEYS[1])
	if not res[1] then
	    return false
	end
	local url = res[1]
	local score = res[2]

	local host = url:match("^https?://([^/]+)")
	if host and redis.call("get", host) == "1" then
	    redis.call("zadd", KEYS[1], score, url)
	    return false
	end

	if redis.call("sismember", KEYS[2], url) == 1 then
	    return false
	end

	return url
	`)

	const maxRetry = 100
	for range maxRetry {
		val, err := script.Run(ctx, Cache, []string{"urls", "visitedUrls"}).Result()
		if err != nil {
			continue
		}

		if valStr, ok := val.(string); ok && valStr != "" {
			return valStr, true, nil
		}

		time.Sleep(10 * time.Millisecond)
	}
	utils.Log.Cache().Debug("failed to get valid URL after retries")
	return "", false, fmt.Errorf("no valid URL after %d retries", maxRetry)
}

// AddUrls adds multiple URLs to Redis set.
// Logs warnings for empty input and errors for failed Redis operations.
func AddUrls(urls []string) error {
	if len(urls) == 0 {
		utils.Log.Cache().Debug("empty URL list passed to AddUrls")
		return nil
	}

	pipe := Cache.Pipeline()

	for _, url := range urls {
		pipe.ZIncrBy(ctx, "urls", 1, url)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		utils.Log.Cache().Debug("add URLs", "error", err)
		return fmt.Errorf("add URLs: %w", err)
	}

	return nil
}

// AddToVisitedUrl adds a URL to the visitedUrls set.
// Logs warnings for empty input and errors if Redis fails.
func AddToVisitedUrl(u string) error {
	if u == "" {
		utils.Log.Cache().Debug("empty URL passed to AddToVisitedUrl")
		return nil
	}

	err := Cache.SAdd(ctx, "visitedUrls", u).Err()
	if err != nil {
		utils.Log.Cache().Debug("add to visited URLs", "url", u, "error", err)
		return fmt.Errorf("add to visited URLs: %w", err)
	}

	return nil
}

// AddToWaitedHost adds a host key in Redis with a TTL corresponding to delay.
// Logs warnings for empty input and errors for Redis failures.
func AddToWaitedHost(h string, delay int) {
	if h == "" {
		utils.Log.Cache().Debug("empty host passed to AddToWaitedHost")
		return
	}

	err := Cache.Set(ctx, h, 1, time.Duration(delay)*time.Second).Err()
	if err != nil {
		utils.Log.Cache().Debug("add to waited host", "host", h, "error", err)
		return
	}
}
