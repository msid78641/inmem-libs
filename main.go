package main

import (
	"fmt"
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
	for i := range 10000 {
		i = (i % 100) + 1
		waitGroup.Add(1)
		go func() {
			val, _ := to_do.ToDoListStore.Get(strconv.Itoa(i))
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
	fmt.Println("Hit Count -> ", hitCounter, " , Miss Count -> ", missCounter)
}

func main() {
	fmt.Println("Fetching the to dos from the cache")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, server is running. Access /debug/pprof/ for profiles."))
	})
	go func() {
		for range 50 {
			fetchToDos()
			time.Sleep(time.Second * 5)
		}
	}()
	fmt.Println("Starting server on :8080...")
	// 2. Start the HTTP server
	log.Fatal(http.ListenAndServe(":8080", nil))
}
