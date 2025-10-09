package store

import "context"

var (
	ctx   context.Context = context.Background()
	Cache                 = NewCacheClient()
)
