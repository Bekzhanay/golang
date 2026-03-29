// The final value is not always 1000 because multiple goroutines
// increment the shared counter concurrently without synchronization,
// causing a data race and lost updates.

package main

import (
	"fmt"
	"sync"
)

func main() {
	var counter int
	var wg sync.WaitGroup

	for i := 0; i < 1000; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			counter++
		}()
	}

	wg.Wait()
	fmt.Println("Unsafe counter:", counter)
}