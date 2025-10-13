package store

import (
	"bytes"
	"context"
	"encoding/gob"
	"log"
	"time"

	"github.com/redis/go-redis/v9"

	"spider/internal/entity"
	"spider/internal/utils"
)

// NewCacheClient initializes and returns a Redis client.
// Registers entity.Host type with gob for serialization.
func NewCacheClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // No password set
		DB:       0,  // Use default DB
		Protocol: 2,  // Connection protocol

	})

	// Ping to ensure Redis connection is alive
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to connected to redis ERROR: %v", err)
	}
	// Required for gob encoding/decoding of Host structs
	gob.Register(entity.Host{})
	return client
}

// AddHostMetaData serializes a Host struct with gob and stores it in Redis.
// Logs warnings for invalid input and errors for encoding/storage failures.
func AddHostMetaData(h string, host *entity.Host) error {
	if h == "" || host == nil {
		utils.Log.Cache().Warn("invalid input for AddHostMetaData", "host", h)
		return nil
	}

	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(host); err != nil {
		utils.Log.Cache().Error("failed to encode host metadata", "error", err, "host", h)

		return err
	}

	err := Cache.HSet(ctx, "hosts", h, buf.Bytes()).Err()
	if err != nil {
		utils.Log.Cache().Error("failed to store host metadata", "error", err, "host", h)
		return err
	}

	utils.Log.Cache().Info("host metadata stored successfully", "host", h)
	return nil
}

// GetHostMetaData retrieves and decodes a Host struct from Redis.
// Logs each step: retrieval, missing host, decode error, successful fetch.
func GetHostMetaData(h string) (*entity.Host, bool) {
	utils.Log.Cache().Info("retrieving host metadata", "host", h)

	val, err := Cache.HGet(ctx, "hosts", h).Bytes()
	if err == redis.Nil {
		utils.Log.Cache().Warn("host metadata not found", "host", h)
		return nil, false
	}
	if err != nil {
		utils.Log.Cache().Error("redis error retrieving host metadata", "error", err, "host", h)
		return nil, false
	}

	var host entity.Host
	if err := gob.NewDecoder(bytes.NewReader(val)).Decode(&host); err != nil {
		utils.Log.Cache().Error("failed to decode host metadata", "error", err, "host", h)
		return nil, false
	}

	utils.Log.Cache().Info("host metadata retrieved successfully", "host", h)
	return &host, true
}

// GetUrl retrieves a URL from Redis atomically.
// Skips URLs if host is recently visited or URL is already visited.
// Uses Lua script to ensure atomicity.
func GetUrl() (u string, ok bool, err error) {
	script := redis.NewScript(`
-- KEYS[1] = urls set
-- KEYS[2] = visitedUrls set

local url = redis.call("zpopmax", KEYS[1])
if not url then
    return nil
end

-- parse host from URL in Lua (simplified: assume full URL)
local host = url:match("^https?://([^/]+)")
if host then
    if redis.call("get", host) == "1" then
        redis.call("sadd", KEYS[1], url)
        return nil
    end
end

if redis.call("sismember", KEYS[2], url) == 1 then
    redis.call("sadd", KEYS[1], url)
    return nil
end

return url
`)

	for {
		val, err := script.Run(ctx, Cache, []string{"urls", "visitedUrls"}).Result()
		// if err == redis.Nil {
		// 	utils.Log.Cache().Info("No URLs left in Redis")
		// 	return "", false, nil
		// }
		if err != nil {
			utils.Log.Cache().Debug("Redis Lua Failed", "error", err)
			// time.Sleep(50 * time.Millisecond) // avoid busy spin
			continue
		}

		if valStr, ok := val.(string); ok && valStr != "" {
			utils.Log.Cache().Info("URL retrieved successfully", "url", valStr)
			return valStr, true, nil
		}

		// skipped URL due to visited host or visited URL
		utils.Log.Cache().Debug("Skipped URL from Lua script, retrying")
		time.Sleep(10 * time.Millisecond)
	}
}

// AddUrls adds multiple URLs to Redis set.
// Logs warnings for empty input and errors for failed Redis operations.
func AddUrls(urls []string) {
	if len(urls) == 0 {
		utils.Log.Cache().Debug("no URLs provided to AddUrls")
		return
	}
	count := 0
	for _, url := range urls {
		err := Cache.ZIncrBy(ctx, "urls", 1, url).Err()
		if err != nil {
			utils.Log.Cache().Warn("failed to add URLs", "error", err)
			continue
		}
		count++
	}
	utils.Log.Cache().Info("URLs added successfully", "count", count)
}

// AddToVisitedUrl adds a URL to the visitedUrls set.
// Logs warnings for empty input and errors if Redis fails.
func AddToVisitedUrl(u string) {
	if u == "" {
		utils.Log.Cache().Debug("empty URL passed to AddToVisitedUrl")
		return
	}
	err := Cache.SAdd(ctx, "visitedUrls", u).Err()
	if err != nil {
		utils.Log.Cache().Warn("failed to add visited URL", "url", u, "error", err)
		return
	}
	utils.Log.Cache().Info("added visited URL", "url", u)
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
		utils.Log.Cache().Warn("failed to add waited host", "host", h, "error", err)
		return
	}
	utils.Log.Cache().Info("added waited host", "host", h)
}
