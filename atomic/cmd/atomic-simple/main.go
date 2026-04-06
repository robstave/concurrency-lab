package main

import (
	"fmt"
	"sync/atomic"
)

func main() {
	var ops atomic.Uint64
	// Single-goroutine increments using typed atomic counter
	for i := 0; i < 5000; i++ {
		ops.Add(1)
	}

	// looking for 5000 total ops
	fmt.Println("ops:", ops.Load())

}
