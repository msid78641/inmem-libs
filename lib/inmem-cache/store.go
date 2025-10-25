package inmem_cache

import "time"

var UserDetailsStore *Cache = createCache(GetBigCache(), time.Second*12000)

var ProductCache *Cache = createCache(GetBigCache(), time.Second*40)
