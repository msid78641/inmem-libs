package main

import (
	"fmt"
	cache "inmem/lib/inmem-cache"
	to_do "inmem/src/to-do"
	"log"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

func fetchToDos() {
	waitGroup := sync.WaitGroup{}
	var hitCounter atomic.Int32
	var missCounter atomic.Int32
	for i := range 100 {
		i = (i % 100) + 1
		waitGroup.Add(1)
		go func() {
			cacheOptions := []cache.CacheOptions{
				cache.WithByPass(cache.WithLoader(to_do.GetToDoLoader)),
				cache.WithLoader(to_do.GetToDoLoader),
				cache.WithStaleResponse(time.Second*50, cache.WithLoader(to_do.GetToDoLoader)),
			}
			val, _ := to_do.ToDoListStore.Get(strconv.Itoa(i), cacheOptions...)
			if val != nil {
				hitCounter.Add(1)
				fmt.Println("Prinitng results for the ", i, " -> hit")
			} else {
				missCounter.Add(1)
				fmt.Println("Prinitng results for the ", i, " -> miss")
			}
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
	fmt.Println("Hit Count -> ", hitCounter.Load(), " , Miss Count -> ", missCounter.Load())
}

func main() {
	fmt.Println("Fetching the to dos from the cache")
	go func() {
		for range 50 {
			fetchToDos()
			time.Sleep(time.Second * 1)
		}
	}()
	fmt.Println("Starting server on :8080...")
	// 2. Start the HTTP server
	log.Fatal(http.ListenAndServe(":8081", nil))
}
