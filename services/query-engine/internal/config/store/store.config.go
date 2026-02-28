package store

import "query-engine/internal/util"

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
