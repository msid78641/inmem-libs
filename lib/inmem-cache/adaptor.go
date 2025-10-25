package inmem_cache

type CacheAdaptorServiceContract interface {
	Load(key string) (*CacheEntry, error)
	Set(key string, cacheEntry *CacheEntry) error
	Delete(key string) error
	SoftDelete(key string) error
}
