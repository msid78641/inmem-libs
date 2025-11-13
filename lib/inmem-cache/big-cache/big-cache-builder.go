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

type OptionalBigCacheConfig func(b *bigcache.Config)

func WithShards(shards int) OptionalBigCacheConfig {
	return func(b *bigcache.Config) {
		b.Shards = shards
	}
}

func WithEnableStats(enableStats bool) OptionalBigCacheConfig {
	return func(b *bigcache.Config) {
		b.StatsEnabled = enableStats
	}
}

func WithCacheMemoryLimit(cacheMemoryLimit int) OptionalBigCacheConfig {
	return func(b *bigcache.Config) {
		b.HardMaxCacheSize = cacheMemoryLimit
	}
}
func WithVerbose(verbose bool) OptionalBigCacheConfig {
	return func(b *bigcache.Config) {
		b.Verbose = verbose
	}
}

func CreateBigCache(optionalBigCacheConfigs ...OptionalBigCacheConfig) *BigCacheAdapter {
	cfg := bigcache.Config{
		Shards:           4,
		LifeWindow:       time.Second * 1000,
		StatsEnabled:     false,
		Verbose:          false,
		HardMaxCacheSize: 1000,
	}
	for _, option := range optionalBigCacheConfigs {
		option(&cfg)
	}
	bigCache, _ := bigcache.New(context.Background(), cfg)
	return &BigCacheAdapter{
		cache: bigCache,
	}
}
