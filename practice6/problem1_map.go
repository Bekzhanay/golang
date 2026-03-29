package main

import (
	"fmt"
	"sync"
)

func syncMapExample() {
	var m sync.Map
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			m.Store("key", val)
		}(i)
	}

	wg.Wait()

	value, _ := m.Load("key")
	fmt.Printf("sync.Map value: %v\n", value)
}

func rwMutexMapExample() {
	regularMap := make(map[string]int)
	var mu sync.RWMutex
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			mu.Lock()
			regularMap["key"] = val
			mu.Unlock()
		}(i)
	}

	wg.Wait()

	mu.RLock()
	value := regularMap["key"]
	mu.RUnlock()

	fmt.Printf("RWMutex map value: %d\n", value)
}

func main() {
	fmt.Println("Problem 1: Thread-safe maps")
	syncMapExample()
	rwMutexMapExample()
}
