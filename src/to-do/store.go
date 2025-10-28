package to_do

import (
	cache "inmem/lib/inmem-cache"
	big_cache "inmem/lib/inmem-cache/big-cache"
	"time"
)

const (
	CacheTTL = time.Second * 3
)

var optionalBigCacheConfigs = []big_cache.OptionalBigCacheConfig{
	big_cache.WithShards(4),
	big_cache.WithCacheMemoryLimit(2),
}

var bigCache = big_cache.CreateBigCache(optionalBigCacheConfigs...)

var toDoListCacheOptions = []cache.OptionalCacheConfigFunc{
	cache.WithLoader(GetToDoLoader),
	cache.WithStaleResponse(time.Second * 5),
}
var ToDoListStore = cache.GetCache(bigCache, CacheTTL, toDoListCacheOptions...)
