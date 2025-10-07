package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// Demonstrates atomic.Uint64 typed API with multiple goroutines performing
// a fixed number of increments each.
func main() {
	var ops atomic.Uint64
	var wg sync.WaitGroup

	workers := 50
	iters := 1000

	wg.Add(workers)
	for range workers {
		go func() {
			defer wg.Done()
			for range iters {
				ops.Add(1)
			}
		}()
	}

	wg.Wait()

	// looking for 50,000 total ops (50 workers * 1000 iters each)
	fmt.Println("ops:", ops.Load())
}
