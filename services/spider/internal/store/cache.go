package store

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/redis/go-redis/v9"
)

type CacheClient struct{ *redis.Client }

var (
	ctx   context.Context = context.Background()
	Cache                 = NewCacheClient()
)

var mu sync.Mutex

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

func PopHost() (url string, err error) {
	script := redis.NewScript(`
local host = redis.call('spop', KEYS[1])
if not host then
    return nil
end
if redis.call('sismember', KEYS[3], host) == 1 or redis.call('hexists', KEYS[2], host) == 1 then
    return nil
end
redis.call('hset', KEYS[2], host, 1)
return host
`)

	host, err := script.Run(ctx, Cache.Client, []string{"newHost", "inProg", "complHosts"}).Result()
	if err != nil {
		return "", err
	}
	if host == nil {
		return "", fmt.Errorf("no host available")
	}

	return host.(string), nil
}

func InProg(host string) (err error) {
	ok, err := Cache.HSetNX(ctx, "inProg", host, 1).Result()
	if err != nil {
		return
	}
	if !ok {
		return fmt.Errorf("Host already in progress")
	}
	return nil
}

func Completed(host string) error {
	script := redis.NewScript(`
local del = redis.call('hdel', KEYS[1], KEYS[3])
if del == 0 then
	return 0
end

local add = redis.call('sadd', KEYS[2], KEYS[3])
return 1
`)

	res, err := script.Run(ctx, Cache.Client, []string{"inProg", "complHosts", host}).Result()
	if err != nil {
		return err
	}

	val, ok := res.(int64)
	if !ok || val == 0 {
		return fmt.Errorf("Failed to cleanUp host: %s", host)
	}

	return nil
}

func AddLink(host, links string) (err error) {
	val, err := Cache.SAdd(ctx, host, links).Result()
	if err != nil {
		return
	}
	if val == 0 {
		return fmt.Errorf("Link already added")
	}
	return
}

func AddHost(host string) (err error) {
	val, err := Cache.SAdd(ctx, "Hosts", host).Result()
	if err != nil {
		return
	}
	if val == 0 {
		return fmt.Errorf("Host already added")
	}
	return
}
