package config

import (
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type PsqlConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string

	MaxConns int
	MaxIdle  int

	MaxOpenConns int
	MaxIdleConns int
}

type StoreConfig struct {
	DB PsqlConfig
}

type RankingConfig struct {
	MaxResults int
	WeightTF   float64
}

type Config struct {
	Store  StoreConfig
	Ranker RankingConfig
}

func LoadConfig(envPath string) (*Config, error) {
	err := godotenv.Load(envPath)
	if err != nil {
		return nil, fmt.Errorf("Error loading .env file")
	}

	c := &Config{
		Ranker: LoadRankingConfig(),
		Store:  loadStoreConfig(),
	}

	return c, nil
}

func loadStoreConfig() StoreConfig {
	return StoreConfig{
		DB: loadDatabaseConfig(),
	}
}

func LoadRankingConfig() RankingConfig {
	maxResults := getIntWithDefault("RANKER_MAX_RESULTS", 100)
	weightTF := getFloatWithDefault("RANKER_WEIGHT_TF", 0.5)

	return RankingConfig{
		maxResults,
		weightTF,
	}
}

func loadDatabaseConfig() PsqlConfig {
	host := getWithDefault("PG_HOST", "localhost")
	port := getIntWithDefault("PG_PORT", 5432)
	user := getWithDefault("PG_USER", "admin")
	password := getWithDefault("PG_PASSWORD", "1234")
	dbname := getWithDefault("PG_DBNAME", "se")
	maxOpenConns := getIntWithDefault("PG_MAX_OPEN_CONNS", 20)
	maxIdleConns := getIntWithDefault("PG_MAX_IDLE_CONNS", 20)

	return PsqlConfig{
		Host:     host,
		Port:     port,
		User:     user,
		DBName:   dbname,
		Password: password,

		MaxOpenConns: maxOpenConns,
		MaxIdleConns: maxIdleConns,
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

func getFloatWithDefault(key string, defaultValue float64) float64 {
	k := getWithDefault(key, "")
	v, err := strconv.ParseFloat(k, 64)
	if err != nil {
		return defaultValue
	}
	return v
}
