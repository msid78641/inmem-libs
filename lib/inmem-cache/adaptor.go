package inmem_cache

type CacheAdaptorServiceContract interface {
	Get(key string) (*CacheEntry, error)
	Set(key string, cacheEntry *CacheEntry) error
	Delete(key string) error
}
