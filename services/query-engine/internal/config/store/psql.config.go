package store

import "query-engine/internal/util"

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

func NewDatabaseConfig() PsqlConfig {
	host := util.GetWithDefault("PG_HOST", "localhost")
	port := util.GetIntWithDefault("PG_PORT", 5432)
	user := util.GetWithDefault("PG_USER", "admin")
	password := util.GetWithDefault("PG_PASSWORD", "1234")
	dbname := util.GetWithDefault("PG_DBNAME", "se")
	maxOpenConns := util.GetIntWithDefault("PG_MAX_OPEN_CONNS", 20)
	maxIdleConns := util.GetIntWithDefault("PG_MAX_IDLE_CONNS", 20)

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
