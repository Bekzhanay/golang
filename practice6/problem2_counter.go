package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

func mutexCounter() {
	var counter int
	var mu sync.Mutex
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mu.Lock()
			counter++
			mu.Unlock()
		}()
	}

	wg.Wait()
	fmt.Println("Mutex counter:", counter)
}

func atomicCounter() {
	var counter int64
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddInt64(&counter, 1)
		}()
	}

	wg.Wait()
	fmt.Println("Atomic counter:", counter)
}

func main() {
	fmt.Println("Problem 2: Concurrent Counter")
	mutexCounter()
	atomicCounter()
}