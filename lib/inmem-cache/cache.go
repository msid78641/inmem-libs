package inmem_cache

import (
	"errors"
	"fmt"
	"golang.org/x/sync/singleflight"
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

type DeletionResult struct {
	Success []string
	Failed  []error
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
	defer func() {
		if err != nil {
			err = cacheError(GET, key, err)
		}
	}()
	optionalConfig := getCacheOptions(options)
	if optionalConfig.bypass {
		if optionalConfig.loader == nil {
			return nil, ErrLoaderNil
		}
		return c.load(key, optionalConfig.loader)
	}
	val, err := c.cacheAdaptor.Get(key)
	if err != nil {
		if !errors.Is(err, ErrEntryNotFound) {
			return nil, err
		}
		if optionalConfig.loader == nil {
			return nil, ErrEntryNotFound
		}
		fmt.Println("Entry not found loading from the loader key -> ", key)
		return c.loadAndSet(key, optionalConfig.loader)
	} else if val.isInValidEntry(0) {
		if optionalConfig.loader == nil || val.isInValidEntry(optionalConfig.staleResponseTtl) {
			c.Delete(DeleteWithKeys([]string{key}))
			return nil, ErrStaleResponse
		}
		fmt.Println("Entry is invalid serving stale response with key -> ", key)
		// here since the serve stale is set we should return the stale response but also load the value in background
		return c.loadAndSet(key, optionalConfig.loader)
	}
	return val.Value, nil
}

func (c *Cache) loadAndSet(key string, loader loaderContract) (interface{}, error) {
	newVal, err := c.load(key, loader)
	if err == nil {
		return nil, ErrLoaderFailed
	}
	c.Set(key, newVal)
	return newVal, nil
}
func (c *Cache) load(key string, loader loaderContract) (interface{}, error) {
	//Only fetch a key once; if already being fetched, block other goroutines until the fetch completes.
	v, err, _ := c.loaderGroup.Do(key, func() (interface{}, error) {
		fmt.Println("Making use of loader for the key ", key)
		val, err := loader(key)
		return val, err
	})
	return v, err
}

func (c *Cache) Set(key string, val any, keyTags ...string) (err error) {
	defer func() {
		if err != nil {
			err = cacheError(SET, key, err)
		}
	}()
	err = c.setKeyValueWithCustomTtl(key, val, c.ttl)
	if err == nil {
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
func (c *Cache) Delete(deleteOpts ...DeleteOptions) (deletionRes *DeletionResult, err error) {
	defer func() {
		if err != nil {
			err = cacheError(DELETE, "", err)
		}
	}()
	deleteConfig := getDeleteOptionConfig(deleteOpts)
	keys := []string{}
	var deletionError error
	deletionRes = &DeletionResult{
		Failed:  []error{},
		Success: []string{},
	}
	if len(deleteConfig.keys) > 0 {
		keys = deleteConfig.keys
		c.deleteThreshold.Add(1)
	} else if len(deleteConfig.tags) > 0 {
		keys = c.getKeysByTag(deleteConfig.tags)
		fmt.Println("Deleteing below keys because of the tag map keys ", keys)
	} else {
		return nil, ErrInvalidDeletionArgs
	}
	for _, key := range keys {
		err = c.cacheAdaptor.Delete(key)
		if err != nil {
			cacheError := &CacheError{
				Operation: DELETE,
				Key:       key,
				BaseError: err,
			}
			deletionError = errors.Join(deletionError, cacheError)
			deletionRes.Failed = append(deletionRes.Failed, cacheError)
		} else {
			deletionRes.Success = append(deletionRes.Success, key)
		}
	}
	if c.deleteThreshold.Load() > 20 {
		c.loaderGroup.Do("cleanup", func() (interface{}, error) {
			c.cleanupMapOnThreshold()
			c.deleteThreshold.Store(0)
			return nil, nil
		})
		if err != nil {
			return nil, err
		}
	}
	return deletionRes, deletionError
}
func (c *Cache) SoftDelete(key string) (err error) {
	defer func() {
		if err != nil {
			err = cacheError(SOFTDELETE, "", err)
		}
	}()
	val, err := c.Get(key)
	if err != nil {
		if errors.Is(err, ErrEntryNotFound) {
			return ErrEntryNotFound
		}
		return err
	}
	fmt.Println("Soft deleting the entry for key ", key)
	return c.setKeyValueWithCustomTtl(key, val, 0)
}

func (c *Cache) setKeyValueWithCustomTtl(key string, value interface{}, ttl time.Duration) error {
	cacheEntry := &CacheEntry{Value: value, TTL: time.Duration(time.Now().Add(ttl).UnixNano())}
	return c.cacheAdaptor.Set(key, cacheEntry)
}

func (c *Cache) getKeysByTag(tags []string) []string {
	fmt.Println(c.tags)
	keys := []string{}
	for _, tag := range tags {
		keys = append(keys, c.tags[tag]...)
	}
	return keys
}

func (c *Cache) cleanupMapOnThreshold() {
	fmt.Println("Mock clean up woke up")
	defer c.tagsMutex.Unlock()
	c.tagsMutex.Lock()
	for tagKey, keys := range c.tags {
		updatedKeys := []string{}
		for _, key := range keys {
			_, err := c.Get(key)
			if !errors.Is(err, ErrEntryNotFound) {
				updatedKeys = append(updatedKeys, key)
			}
		}
		if len(updatedKeys) > 0 {
			c.tags[tagKey] = updatedKeys
		} else {
			delete(c.tags, tagKey)
		}
	}
	fmt.Println("Mock clean up close ")
}
