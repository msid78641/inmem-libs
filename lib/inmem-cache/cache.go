package inmem_cache

import (
	"fmt"
	"golang.org/x/sync/singleflight"
	"strings"
	"time"
)

type loaderContract func(key string) (interface{}, error)
type Cache struct {
	cacheAdaptor CacheAdaptorServiceContract
	ttl          time.Duration
	loaderGroup  singleflight.Group
}

type CacheOptions func(c *cacheOptionsConfig)

type cacheOptionsConfig struct {
	loader           loaderContract
	staleResponseTtl time.Duration
	bypass           bool
}

func WithLoader(loader loaderContract) CacheOptions {
	return func(c *cacheOptionsConfig) {
		c.loader = loader
	}
}

func WithStaleResponse(staleTtl time.Duration, options ...CacheOptions) CacheOptions {
	return func(c *cacheOptionsConfig) {
		c.staleResponseTtl = staleTtl
		if len(options) > 0 {
			for _, option := range options {
				option(c)
			}
		}
	}
}

func WithByPass(options ...CacheOptions) CacheOptions {
	return func(c *cacheOptionsConfig) {
		c.bypass = true
		if len(options) > 0 {
			for _, option := range options {
				option(c)
			}
		}
	}
}

func GetCache(cacheAdaptor CacheAdaptorServiceContract, ttl time.Duration) *Cache {
	newCacheWithDefaultConfig := &Cache{
		cacheAdaptor: cacheAdaptor,
		ttl:          ttl,
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

func getCacheOptions(options []CacheOptions) *cacheOptionsConfig {
	var optionalConfig *cacheOptionsConfig = &cacheOptionsConfig{}
	for _, option := range options {
		option(optionalConfig)
	}
	return optionalConfig
}
func (c *Cache) Get(key string, options ...CacheOptions) (interface{}, error) {
	optionalConfig := getCacheOptions(options)
	if optionalConfig.bypass {
		return c.load(key, optionalConfig.loader)
	} else {
		val, err := c.cacheAdaptor.Get(key)
		if err != nil {
			fmt.Println(err.Error())
			isEntryNotFoundError := strings.Compare(err.Error(), "Entry not found") == 0
			if isEntryNotFoundError {
				fmt.Println("Entry not found loading from the loader key -> ", key)
				// if loader is present in the option else return the error
				if optionalConfig.loader != nil {
					val, err := c.load(key, optionalConfig.loader) // here since the entry is not found load it from live and send
					if err == nil {
						err := c.Set(key, val)
						if err != nil {
							fmt.Println("Setting error ", err)
							return nil, err
						}
					}
					return val, err
				}
				return nil, nil
			}
			fmt.Println("Some error occurred while fetching from the in mem cache ", err)
			return nil, err
		} else if val.isInValidEntry(0) {
			if val.isInValidEntry(optionalConfig.staleResponseTtl) {
				fmt.Println("Deleting the invalid entry key -> ", key)
				c.Delete(key)
				return nil, nil
			}
			fmt.Println("Entry is invalid serving stale response with key -> ", key)
			// here since the serve stale is set we should return the stale response but also load the value in background
			newVal, err := c.load(key, optionalConfig.loader)
			if err != nil {
				c.Set(key, newVal)
			}
			return val.Value, nil
		}
		return val.Value, nil
	}
}

func (c *Cache) load(key string, loader loaderContract) (interface{}, error) {
	//Only fetch a key once; if already being fetched, block other goroutines until the fetch completes.
	v, err, _ := c.loaderGroup.Do(key, func() (interface{}, error) {
		fmt.Println("Making use of loader for the key ", key)
		val, err := loader(key)
		if err != nil {
			fmt.Println("Error resulted from the loader, loading the key ", key)
			return nil, nil
		}
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
