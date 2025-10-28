package inmem_cache

import (
	"fmt"
	"golang.org/x/sync/singleflight"
	"strings"
	"time"
)

type loaderContract func(key string) (interface{}, error)
type Cache struct {
	cacheAdaptor     CacheAdaptorServiceContract
	ttl              time.Duration
	loader           loaderContract
	staleResponseTtl time.Duration
	loaderGroup      singleflight.Group
}

type optionalCacheConfig struct {
	loader        loaderContract
	staleResponse time.Duration
}

type OptionalCacheConfigFunc func(c *Cache)

func WithLoader(loader loaderContract) OptionalCacheConfigFunc {
	return func(c *Cache) {
		c.loader = loader
	}
}

func WithStaleResponse(staleTtl time.Duration) OptionalCacheConfigFunc {
	return func(c *Cache) {
		c.staleResponseTtl = staleTtl
	}
}
func GetCache(cacheAdaptor CacheAdaptorServiceContract, ttl time.Duration, options ...OptionalCacheConfigFunc) *Cache {
	newCacheWithDefaultConfig := &Cache{
		cacheAdaptor:     cacheAdaptor,
		ttl:              ttl,
		staleResponseTtl: 0,
	}
	for _, opt := range options {
		opt(newCacheWithDefaultConfig)
	}
	return newCacheWithDefaultConfig
}

type CacheEntry struct {
	Value interface{}
	TTL   time.Duration
}

func (ce *CacheEntry) isInValidEntry(buffer time.Duration) bool {
	if ce.TTL <= time.Duration(time.Now().Add(-buffer).UnixNano()) {
		return true
	}
	return false
}
func (c *Cache) Get(key string) (interface{}, error) {
	val, err := c.cacheAdaptor.Get(key)
	if err != nil {
		isEntryNotFoundError := strings.Compare(err.Error(), "Entry not found") == 0
		if isEntryNotFoundError {
			fmt.Println("Entry not found loading from the loader key -> ", key)
			return c.Load(key) // here since the entry is not found load it from live and send
		}
		fmt.Println("Some error occurred while fetching from the in mem cache ", err)
		return nil, err
	} else if val.isInValidEntry(0) {
		if val.isInValidEntry(c.staleResponseTtl) {
			fmt.Println("Deleting the invalid entry key -> ", key)
			c.Delete(key)
			return nil, nil
		}
		fmt.Println("Entry is invalid serving stale response with key -> ", key)
		go c.Load(key) // here since the serve stale is set we should return the stale response but also load the value in background
		return val.Value, nil
	}
	return val.Value, nil
}

func (c *Cache) Load(key string) (interface{}, error) {
	// Circuit breaker if the key is already been fetched do not fetch it again instead
	v, err, _ := c.loaderGroup.Do(key, func() (interface{}, error) {
		fmt.Println("Making use of loader for the key ", key)
		val, err := c.loader(key)
		if err != nil {
			fmt.Println("Error resulted from the loader, loading the key ", key)
			return nil, nil
		}
		c.Set(key, val)
		return val, err
	})
	return v, err
}

func (c *Cache) Set(key string, val interface{}) error {
	return c.setKeyValueWithCustomTtl(key, val, c.ttl)
}

func (c *Cache) Delete(key string) error {
	err := c.cacheAdaptor.Delete(key)
	if err != nil {
		// handle error
	}
	return nil
}
func (c *Cache) SoftDelete(key string) error {
	val, err := c.Get(key)
	if err != nil {
		// handle error
	}
	return c.setKeyValueWithCustomTtl(key, val, 0)
}

func (c *Cache) setKeyValueWithCustomTtl(key string, value interface{}, ttl time.Duration) error {
	cacheEntry := &CacheEntry{Value: value, TTL: time.Duration(time.Now().Add(ttl).UnixNano())}
	err := c.cacheAdaptor.Set(key, cacheEntry)
	if err != nil {
		// handle error
	}
	return nil
}
