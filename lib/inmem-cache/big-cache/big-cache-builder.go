package big_cache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"time"
)

type BigCacheConfig struct {
	shards           int
	lifeWindow       time.Duration
	enableStats      bool
	verbose          bool
	cacheMemoryLimit int
}

func CreateBigCache() *BigCacheConfig {
	return &BigCacheConfig{
		shards:           4,
		lifeWindow:       -1,
		enableStats:      false,
		verbose:          false,
		cacheMemoryLimit: 10,
	}
}

func (b *BigCacheConfig) Shards(shards int) *BigCacheConfig {
	b.shards = shards
	return b
}
func (b *BigCacheConfig) LifeWindow(lifeWindow time.Duration) *BigCacheConfig {
	b.lifeWindow = lifeWindow
	return b
}
func (b *BigCacheConfig) EnableStats(enableStats bool) *BigCacheConfig {
	b.enableStats = enableStats
	return b
}
func (b *BigCacheConfig) Verbose(verbose bool) *BigCacheConfig {
	b.verbose = verbose
	return b
}

func (b *BigCacheConfig) CacheMemoryLimit(memoryLimit int) *BigCacheConfig {
	b.cacheMemoryLimit = memoryLimit
	return b
}

func (b *BigCacheConfig) Build() *BigCacheAdapter {
	bigCache, _ := bigcache.New(context.Background(), bigcache.Config{
		Shards:           b.shards,
		LifeWindow:       b.lifeWindow, // BigCache doesn't evict with jitter. The caching layer must implement TTL to add jitter
		StatsEnabled:     b.enableStats,
		Verbose:          b.verbose,
		HardMaxCacheSize: b.cacheMemoryLimit,
	})
	return &BigCacheAdapter{
		cache: bigCache,
	}
}
