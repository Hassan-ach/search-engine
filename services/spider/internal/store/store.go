package store

import (
	"context"

	"spider/internal/config"
	"spider/internal/entity"
)

type Store struct {
	DB    DB
	Cache Cache
}

type Cache interface {
	AddHostMetaData(h string, host *entity.Host) error
	GetHostMetaData(h string) (*entity.Host, bool, error)
	GetUrl() (string, bool, error)
	AddUrls(urls []string) error
	AddToVisitedUrl(u string) error
	AddToWaitedHost(h string, delay int)
}
type DB interface {
	InsertPage(page entity.Page) error
	InsertGraphEdges(from_url_id string, to_url_ids []string) error
	InsertURLs(urls []string) ([]string, error)
}

func NewStore(ctx context.Context, conf *config.Config) *Store {
	db := NewDbClient(ctx, conf.DB)
	rd := NewRedisClient(ctx, conf.Redis)
	return &Store{
		DB:    db,
		Cache: rd,
	}
}
