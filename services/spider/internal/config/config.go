package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Port     int
	Delay    int
	MaxRetry int
}

type PSQLConfig struct {
	Host            string
	Port            int
	User            string
	DBname          string
	Password        string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
}

type StoreConfig struct {
	Cache RedisConfig
	DB    PSQLConfig
}

type AppConfig struct {
	MaxWorkers         int
	MaxConcurrentFetch int
	LogsPath           string
	ClawlerDelay       int
	HttpTimeout        int
	CrawlerTimeout     int
}

type Config struct {
	App   AppConfig
	Store StoreConfig
}

func LoadConfig(envPath string) (*Config, error) {
	err := godotenv.Load("../../.env")
	if err != nil {
		panic("Failed to load .env file")
	}

	c := &Config{
		App:   loadAppConfig(),
		Store: loadStoreConfig(),
	}

	fmt.Printf("%+v\n", c)

	return c, nil
}

func loadStoreConfig() StoreConfig {
	return StoreConfig{
		Cache: loadRedisConfig(),
		DB:    loadDatabaseConfig(),
	}
}

func loadDatabaseConfig() PSQLConfig {
	host := getWithDefault("PG_HOST", "localhost")
	port := getIntWithDefault("PG_PORT", 5432)
	user := getWithDefault("PG_USER", "admin")
	password := getWithDefault("PG_PASSWORD", "1234")
	dbname := getWithDefault("PG_DBNAME", "se")
	maxOpenConns := getIntWithDefault("PG_MAX_OPEN_CONNS", 20)
	maxIdleConns := getIntWithDefault("PG_MAX_IDLE_CONNS", 20)
	maxConnLifetime := getIntWithDefault("PG_MAX_CONN_LIFETIME", 0)

	return PSQLConfig{
		Host:            host,
		Port:            port,
		User:            user,
		DBname:          dbname,
		Password:        password,
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		MaxConnLifetime: time.Second * time.Duration(maxConnLifetime),
	}
}

func loadRedisConfig() RedisConfig {
	addr := getWithDefault("REDIS_ADDR", "localhost")
	password := getWithDefault("REDIS_PASSWORD", "")
	db := getIntWithDefault("REDIS_DB", 1)
	port := getIntWithDefault("REDIS_PORT", 6379)
	delay := getIntWithDefault("REDIS_DELAY", 5)
	maxRetry := getIntWithDefault("REDIS_MAX_RETRY", 10)

	return RedisConfig{
		Addr:     addr,
		Password: password,
		Port:     port,
		DB:       db,
		Delay:    delay,
		MaxRetry: maxRetry,
	}
}

func loadAppConfig() AppConfig {
	maxWorkers := getIntWithDefault("MAX_WORKERS", 10)
	httpTimeout := getIntWithDefault("HTTP_TIMEOUT", 10)
	crawlerTimeout := getIntWithDefault("CRAWLER_TIMEOUT", 30)
	maxConcurrentFetch := getIntWithDefault("MAX_CONCURRENT_FETCH", 100)
	logsPath := getWithDefault("LOGS_PATH", "./logs")
	clawlerDelay := getIntWithDefault("CRAWLER_DELAY", 50)
	return AppConfig{
		MaxWorkers:         maxWorkers,
		CrawlerTimeout:     crawlerTimeout,
		HttpTimeout:        httpTimeout,
		MaxConcurrentFetch: maxConcurrentFetch,
		LogsPath:           logsPath,
		ClawlerDelay:       clawlerDelay,
	}
}

func getWithDefault(key, defaultValue string) string {
	k := os.Getenv(key)
	if k == "" {
		return defaultValue
	}
	return k
}

func getIntWithDefault(key string, defaultValue int) int {
	k := getWithDefault(key, "")
	v, err := strconv.Atoi(k)
	if err != nil {
		return defaultValue
	}
	return v
}
