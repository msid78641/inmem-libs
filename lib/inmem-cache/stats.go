package inmem_cache

import (
	"fmt"
	"inmem/lib/logger"
	"sync/atomic"
	"time"
)

type CacheStats struct {
	hit              atomic.Int32
	miss             atomic.Int32
	evictions        atomic.Int32
	staleServe       atomic.Int32
	entriesCount     atomic.Int32
	loadCount        atomic.Int32
	loadTime         atomic.Int64
	memoryUsage      atomic.Int64
	tagInvalidations atomic.Int32
	deleteHits       atomic.Int32
	deleteMisses     atomic.Int32
}

func (c *CacheStats) Hit() {
	c.hit.Add(1)
}
func (c *CacheStats) Miss() {
	c.miss.Add(1)
}
func (c *CacheStats) Evict() {
	c.evictions.Add(1)
}
func (c *CacheStats) Stale() {
	c.staleServe.Add(1)
}
func (c *CacheStats) EntriesCount() {
	c.entriesCount.Add(1)
}
func (c *CacheStats) LoadCount() {
	c.loadCount.Add(1)
}
func (c *CacheStats) LoadTime(time time.Duration) {
	c.loadTime.Add(int64(time.Milliseconds()))
}
func (c *CacheStats) MemoryUsage(memUsage int64) {
	c.memoryUsage.Add(memUsage)
}
func (c *CacheStats) InvalidateTag() {
	c.tagInvalidations.Add(1)
}
func (c *CacheStats) DeleteHit() {
	c.deleteHits.Add(1)
}
func (c *CacheStats) DeleteMiss() {
	c.deleteMisses.Add(1)
}

func (c *CacheStats) Reset() {
	c.hit.Store(0)
	c.miss.Store(0)
	c.evictions.Store(0)
	c.staleServe.Store(0)
	c.entriesCount.Store(0)
	c.loadCount.Store(0)
	c.loadTime.Store(0)
	c.memoryUsage.Store(0)
	c.tagInvalidations.Store(0)
	c.deleteHits.Store(0)
	c.deleteMisses.Store(0)
}

func InitStats() *CacheStats {
	cacheStats := new(CacheStats)
	go cacheStats.LogStats()
	return cacheStats
}

func (c *CacheStats) LogStats() {
	tick := time.Tick(time.Second * 5)
	for range tick {
		hits := c.hit.Load()
		misses := c.miss.Load()
		totalRequests := hits + misses

		deleteHits := c.deleteHits.Load()
		deleteMisses := c.deleteMisses.Load()
		totalDeletes := deleteHits + deleteMisses

		totalEntriesCount := c.entriesCount.Load()
		loadCount := c.loadCount.Load()
		totalLoadTime := c.loadTime.Load()

		liveEntries := totalEntriesCount - deleteHits

		// Hit Ratio
		var hitRatio float64
		if totalRequests > 0 {
			hitRatio = (float64(hits) / float64(totalRequests)) * 100
		}

		// Delete Hit Ratio
		var deleteHitRatio float64
		if totalDeletes > 0 {
			deleteHitRatio = (float64(deleteHits) / float64(totalDeletes)) * 100
		}

		// Avg Load Time
		var avgLoadTime float64
		if loadCount > 0 {
			avgLoadTime = float64(totalLoadTime) / float64(loadCount)
		}

		fields := map[string]string{
			"hits":      fmt.Sprintf("%d", hits),
			"misses":    fmt.Sprintf("%d", misses),
			"hit_ratio": fmt.Sprintf("%.2f", hitRatio),

			"delete_hits":      fmt.Sprintf("%d", deleteHits),
			"delete_misses":    fmt.Sprintf("%d", deleteMisses),
			"delete_hit_ratio": fmt.Sprintf("%.2f", deleteHitRatio),

			"total_load_time": fmt.Sprintf("%d", totalLoadTime),
			"load_count":      fmt.Sprintf("%d", loadCount),
			"avg_load_time":   fmt.Sprintf("%.4f", avgLoadTime),

			"live_entries":  fmt.Sprintf("%d", liveEntries),
			"total_entries": fmt.Sprintf("%d", totalEntriesCount),

			"evictions":         fmt.Sprintf("%d", c.evictions.Load()),
			"tag_invalidations": fmt.Sprintf("%d", c.tagInvalidations.Load()),
			"stale_served":      fmt.Sprintf("%d", c.staleServe.Load()),
		}

		logger.Dispatch(logger.DEBUG, logger.WithEntry().
			WithFieldMap(fields).
			WithMessage("cache stats"))
	}
}
