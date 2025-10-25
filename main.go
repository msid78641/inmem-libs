package main

import (
	"fmt"
	to_do "inmem/src/to-do"
	"strconv"
	"sync"
	"sync/atomic"
)

func fetchToDos() {
	waitGroup := sync.WaitGroup{}
	var hitCounter atomic.Int32
	var missCounter atomic.Int32
	for i := range 1000 {
		i = (i % 10) + 1
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

func setToDos() {
	waitGroup := sync.WaitGroup{}
	for i := range 10 {
		i = i + 1
		waitGroup.Add(1)
		go func() {
			val, err := to_do.GetToDoLoader(strconv.Itoa(i))
			if err != nil {
				return
			}
			to_do.ToDoListStore.Set(strconv.Itoa(i), val)
			waitGroup.Done()
		}()
	}
	waitGroup.Wait()
}
func main() {
	fmt.Println("Fetching the to dos from the cache")
	fetchToDos()
}
