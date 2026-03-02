package config

import (
	"fmt"

	"github.com/Hassan-ach/boogle/services/engine/internal/config/ranker"
	"github.com/Hassan-ach/boogle/services/engine/internal/config/store"

	"github.com/joho/godotenv"
)

type Config struct {
	Store  store.StoreConfig
	Ranker ranker.RankingConfig
}

func LoadConfig(envPath string) (*Config, error) {
	var err error
	if envPath == "" {
		err = godotenv.Load()
	} else {
		err = godotenv.Load(envPath)
	}
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	c := &Config{
		Ranker: ranker.NewRankingConfig(),
		Store:  store.NewStoreConfig(),
	}

	return c, nil
}
