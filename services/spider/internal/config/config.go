package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type PSQLConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBname          string
	MaxOpenConns    int
	MaxIdleConns    int
	MaxConnLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Port     int
	DB       int
	Password string
	Delay    int
	MaxRetry int
}

type StoreConfig struct {
	Cache RedisConfig
	DB    PSQLConfig
}

type AppConfig struct {
	MaxGoRoutines int
	Timeout       time.Duration
}

type Config struct {
	Store StoreConfig
	App   AppConfig
}

func LoadConfig(envPath string) (*Config, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file")
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
	maxGoRoutines := getIntWithDefault("MAX_GO_ROUTINES", 10)
	timeout := getIntWithDefault("TIMEOUT", 30)
	return AppConfig{
		MaxGoRoutines: maxGoRoutines,
		Timeout:       time.Second * time.Duration(timeout),
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
