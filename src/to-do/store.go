package to_do

import (
	cache "inmem/lib/inmem-cache"
	big_cache "inmem/lib/inmem-cache/big-cache"
	"time"
)

var bigCache = big_cache.CreateBigCache().
	Shards(4).
	LifeWindow(-1).
	EnableStats(true).
	Verbose(true).
	CacheMemoryLimit(2).
	Build()

var ToDoListStore = cache.CreateCache(bigCache, time.Second*10)
