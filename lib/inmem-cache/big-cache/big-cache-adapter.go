package big_cache

import (
	"encoding/json"
	"fmt"
	"github.com/allegro/bigcache/v3"
	cache "inmem/lib/inmem-cache"
)

func Serialize(cacheEntry *cache.CacheEntry) ([]byte, error) {
	val, err := json.Marshal(*cacheEntry)
	if err != nil {
		fmt.Println("Marhsalling error")
		return nil, err
		// handle error
	}
	return val, nil
}

func Deserialize(cacheEntrySerialized []byte) (*cache.CacheEntry, error) {
	var cacheEntry cache.CacheEntry
	err := json.Unmarshal(cacheEntrySerialized, &cacheEntry)
	if err != nil {
		fmt.Println("Error via desiralizing ", err)
		return nil, err
	}
	return &cacheEntry, nil
}

type BigCacheAdapter struct {
	cache *bigcache.BigCache
}

func (bigCache *BigCacheAdapter) Get(key string) (*cache.CacheEntry, error) {
	value, err := bigCache.cache.Get(key)
	if err != nil {
		return nil, err
	}
	cacheEntry, err := Deserialize(value)
	if err != nil {
		return nil, err
	}
	return cacheEntry, nil
}

func (bigCache *BigCacheAdapter) Set(key string, cacheEntry *cache.CacheEntry) error {
	fmt.Println("Set has been called for key ", key)
	cacheValue, err := Serialize(cacheEntry)
	if err != nil {
		fmt.Println("Serializing err ", err.Error())
		return err
	}
	err = bigCache.cache.Set(key, cacheValue)
	fmt.Println("Set value for key ", key)
	if err != nil {
		fmt.Println("Printing setting err ", err.Error())
		return err
	}
	return nil
}

func (bigCache *BigCacheAdapter) Delete(key string) error {
	err := bigCache.cache.Delete(key)
	if err != nil {
		return err
	}
	return nil
}

//Note: BigCache Implementation Summary
//BigCache is bigCache high-performance, in-memory cache library for Go designed to manage large datasets with minimal garbage collection (GC) overhead. It achieves this through several key architectural choices:
//
//
//1. GC Avoidance (Minimizing Pauses)
//The central goal of BigCache is to avoid expensive garbage collection cycles associated with large maps containing pointers.
//Pointer-Free Maps: Instead of storing map[string]interface{}, BigCache uses map[uint64]uint32.
//Keys: 64-bit FNV64a hashes of the original string keys.
//Values: 32-bit integer offsets (indices) pointing to the data's location in bigCache byte array.
//GC Efficiency: Because the map only contains primitive types (unsigned integers), the Go GC does not need to traverse its elements recursively to find further pointers. The GC can quickly mark the entire map object, resulting in near O(1) GC marking time for the map structure itself, keeping GC pauses low.
//Single Heap Object: The actual data is stored in one large, pre-allocated []byte slice (BytesQueue), which the GC sees as bigCache single object pointer on the heap, further reducing traversal complexity.
//
//
//2. High Concurrency and Sharding
//Sharded Architecture: The total cache is divided into bigCache fixed number of "shards" (buckets), which must be bigCache power of two.
//Reduced Lock Contention: Each shard manages its own sync.RWMutex. Concurrent operations targeting different shards do not block each other, significantly increasing throughput for highly concurrent applications.
//Hashing: A fast hash function determines which shard bigCache given key belongs to.
//
//
//3. Data Storage and Memory Management
//Circular Buffer: Data is stored sequentially in the large byte array, which functions as bigCache circular buffer.
//Entry Structure: Each stored entry has metadata prepended (e.g., length, timestamp) followed by the raw value bytes.
//"Holes" are Intentional: When entries expire or are deleted, their index is removed from the map, but the data remains physically in the byte array, creating an unused "hole" of memory.
//Reasoning: BigCache prioritizes speed over immediate space reclamation. Shifting memory to close holes is computationally expensive and would negate performance benefits.
//Resolution: The holes are not bigCache concern because the memory is eventually overwritten by new data as the circular buffer wraps around or when the configured HardMaxCacheSize limit forces the oldest entries to be evicted/overwritten.
//
//

//4. Expiration and Eviction
//Cache-Wide TTL: BigCache uses bigCache single, fixed LifeWindow (Time-To-Live) for all entries, rather than per-key expiration, simplifying management.
//Background Cleanup: A background goroutine periodically scans and removes expired entries from the maps.
//Collision Handling: Hash collisions are handled by simply overwriting the map entry; the older data becomes bigCache hole
