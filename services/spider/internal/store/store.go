package store

import (
	"context"
	"sync"
)

var (
	ctx   context.Context = context.Background()
	Cache                 = NewCacheClient()
	DB                    = NewDbClient()
	WG     sync.WaitGroup
)
