package inmem_cache

import (
	"fmt"
	"golang.org/x/sync/singleflight"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type loaderContract func(key string) (interface{}, error)
type Cache struct {
	cacheAdaptor    CacheAdaptorServiceContract
	ttl             time.Duration
	loaderGroup     singleflight.Group
	tags            map[string][]string
	tagsMutex       sync.Mutex
	deleteThreshold atomic.Int32
}

type CacheOptions func(c *cacheOptionsConfig)

type DeleteOptions func(d *deleteOptionsConfig)

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

type deleteOptionsConfig struct {
	tags []string
	keys []string
}

func DeleteWithKeys(keys []string) DeleteOptions {
	return func(d *deleteOptionsConfig) {
		d.keys = keys
	}
}

func DeleteWithTags(tags []string) DeleteOptions {
	return func(d *deleteOptionsConfig) {
		d.tags = tags
	}
}

func getDeleteOptionConfig(delOpts []DeleteOptions) deleteOptionsConfig {
	var opts = deleteOptionsConfig{}
	for _, option := range delOpts {
		option(&opts)
	}
	return opts
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
		tags:         make(map[string][]string),
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
func (c *Cache) Get(key string, options ...CacheOptions) (res interface{}, err error) {
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
				c.Delete()
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

func (c *Cache) Set(key string, val any, keyTags ...string) error {
	err := c.setKeyValueWithCustomTtl(key, val, c.ttl)
	if err != nil {
		for _, tag := range keyTags {
			c.tagsMutex.Lock()
			if c.tags[tag] == nil {
				c.tags[tag] = []string{}
			}
			c.tags[tag] = append(c.tags[tag], key)
			c.tagsMutex.Unlock()
		}
	}
	return err
}
func (c *Cache) Delete(deleteOpts ...DeleteOptions) error {
	deleteConfig := getDeleteOptionConfig(deleteOpts)
	keys := []string{}
	if len(deleteConfig.keys) > 0 {
		keys = deleteConfig.keys
		c.deleteThreshold.Add(1)
	} else if len(deleteConfig.tags) > 0 {
		keys = c.getKeysByTag(deleteConfig.tags)
	}
	for _, key := range keys {
		c.cacheAdaptor.Delete(key)
	}
	if c.deleteThreshold.Load() > 20 {
		go c.cleanupMapOnThreshold()
	}
	return nil
}
func (c *Cache) SoftDelete(key string) error {
	val, err := c.Get(key)
	if err != nil {
		// handle error
	}
	fmt.Println("Soft deleiting the entry for key ", key)
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

func (c *Cache) getKeysByTag(tags []string) []string {
	keys := []string{}
	for _, tag := range tags {
		keys = append(keys, c.tags[tag]...)
	}
	return keys
}

func (c *Cache) cleanupMapOnThreshold() {
	defer c.tagsMutex.Unlock()
	c.tagsMutex.Lock()
	keysToBeDeleted := []string{}
	for _, keys := range c.tags {
		for _, key := range keys {
			_, err := c.Get(key)
			if err != nil {
				keysToBeDeleted = append(keysToBeDeleted, key)
			}
		}
	}
	c.Delete(DeleteWithKeys(keysToBeDeleted))
}
