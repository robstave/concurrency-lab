package main

import (
	"fmt"
	"sync"
	"sync/atomic"
)

// demo increments a shared counter to 1000 using atomics across goroutines.
func main() {
	var counter int64
	var wg sync.WaitGroup

	workers := 2
	incs := int64(1000)

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				// Load-then-CAS to avoid overshooting incs.
				v := atomic.LoadInt64(&counter)
				if v >= incs {
					return
				}
				if atomic.CompareAndSwapInt64(&counter, v, v+1) {
					// successful increment; continue
				}
				// if CAS failed, another goroutine incremented; retry
			}
		}()
	}

	wg.Wait()
	fmt.Printf("final count=%d (expected %d)\n", counter, incs)
}
