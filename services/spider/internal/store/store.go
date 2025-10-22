package store

import (
	"context"
	"sync"
)

var (
	DB                    = NewDbClient()
	ctx   context.Context = context.Background()
	Cache                 = NewCacheClient()
	WG    sync.WaitGroup
)
