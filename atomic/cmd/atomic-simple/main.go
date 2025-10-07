package main

import (
	"fmt"
	"sync/atomic"
)

type MyStruct struct {
	A int
	B int
}

func main() {
	var ops atomic.Uint64
	// Single-goroutine increments using typed atomic counter
	for i := 0; i < 5000; i++ {
		ops.Add(1)
	}

	// looking for 5000 total ops
	fmt.Println("ops:", ops.Load())

	var ops2 atomic.Value
	aValue := &MyStruct{A: 0, B: 1000}

	ops2.Store(aValue)

	// Atomic values are best treated as immutable snapshots. Create a new
	// struct each iteration and store it back, rather than mutating the
	// loaded pointer in place (which would be unsafe if multiple goroutines
	// were involved).
	for i := 0; i < 5000; i++ {
		v := ops2.Load().(*MyStruct)
		next := &MyStruct{A: v.A + 1, B: v.B - 1}
		ops2.Store(next)
	}

	final := ops2.Load().(*MyStruct)
	fmt.Printf("final: A=%d B=%d (expected A=5000 B=-4000)\n", final.A, final.B)
}
