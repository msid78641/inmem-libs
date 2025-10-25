package to_do

import (
	cache "inmem/lib/inmem-cache"
	big_cache "inmem/lib/inmem-cache/big-cache"
	"time"
)

const (
	CacheTTL = time.Second * 3
)

var bigCache = big_cache.CreateBigCache().
	Shards(4).
	LifeWindow(-1).
	EnableStats(false).
	Verbose(false).
	CacheMemoryLimit(2).
	Build()

var ToDoListStore = cache.GetCache(bigCache, CacheTTL).
	WithLoader(GetToDoLoader).
	WithStaleResponseTtl(time.Second * 3)
