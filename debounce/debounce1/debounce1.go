package main

import (
	"fmt"
	"time"
)

type Debouncer struct {
	Tokens chan struct{}
	Quit   chan struct{}
}

// Emitter emits values every 100ms up to 10, then closes the channel.
func Emitter() <-chan int {
	emitChan := make(chan int)
	go func() {
		for i := 1; i <= 10; i++ {
			emitChan <- i
			time.Sleep(100 * time.Millisecond)
		}
		close(emitChan)
	}()
	return emitChan
}

// NewDebouncer creates a debouncer that sends tokens every 300ms.
func NewDebouncer() *Debouncer {
	d := &Debouncer{
		Tokens: make(chan struct{}),
		Quit:   make(chan struct{}),
	}
	go func() {
		for {
			select {
			case <-d.Quit:
				close(d.Tokens)
				return
			case d.Tokens <- struct{}{}:
				time.Sleep(300 * time.Millisecond)
			}
		}
	}()
	return d
}

func main() {
	emitChan := Emitter()
	debouncer := NewDebouncer()
	defer close(debouncer.Quit) // Ensure the debouncer is stopped.

	for val := range emitChan {
		select {
		case <-debouncer.Tokens:
			fmt.Printf("Processed: %d\n", val)
		default:
			fmt.Printf("Dropped: %d\n", val)
		}
	}
}
