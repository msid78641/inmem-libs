package inmem_cache

import (
	"time"
)

type Cache struct {
	cacheAdaptor CacheAdaptorServiceContract
	ttl          time.Duration
}

func CreateCache(cacheAdaptor CacheAdaptorServiceContract, ttl time.Duration) *Cache {
	return &Cache{
		cacheAdaptor: cacheAdaptor,
		ttl:          ttl,
	}
}

type CacheEntry struct {
	Value interface{}
	TTL   time.Duration
}

func (ce *CacheEntry) isValidEntry() bool {
	if ce.TTL >= time.Duration(time.Now().UnixNano()) {
		return true
	}
	return false
}

func (c *Cache) Load(key string) (interface{}, error) {
	val, err := c.cacheAdaptor.Load(key)
	if err != nil {
		// Handle the error
		return nil, err
	}
	if val.isValidEntry() {
		return val.Value, nil
	}
	return nil, nil
}

func (c *Cache) Set(key string, val interface{}) error {
	cacheEntry := &CacheEntry{Value: val, TTL: time.Duration(time.Now().Add(c.ttl).UnixNano())}
	err := c.cacheAdaptor.Set(key, cacheEntry)
	if err != nil {
		// handle error
	}
	return nil
}

func (c *Cache) Delete(key string) error {
	err := c.cacheAdaptor.Delete(key)
	if err != nil {
		// handle error
	}
	return nil
}
func (c *Cache) SoftDelete(key string) error {
	err := c.cacheAdaptor.SoftDelete(key)
	if err != nil {
		// handle error
	}
	return nil
}
