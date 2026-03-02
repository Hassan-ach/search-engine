package store

import "github.com/Hassan-ach/boogle/services/engine/internal/util"

type StoreConfig struct {
	DB       PsqlConfig
	PageSize int
}

func NewStoreConfig() StoreConfig {
	pageSize := util.GetIntWithDefault("PAGE_SIZE", 20)
	return StoreConfig{
		DB:       NewDatabaseConfig(),
		PageSize: pageSize,
	}
}
