package main

import (
	"context"
	"fmt"
	"sync"
)

type request struct {
	ctx   context.Context
	delta int
	reply chan int
}

// Counter's value belongs exclusively to its owner goroutine.
// Callers exchange messages; they never read or write the value directly.
type Counter struct {
	requests chan request
	done     chan struct{}
}

func NewCounter(ctx context.Context) *Counter {
	c := &Counter{requests: make(chan request), done: make(chan struct{})}
	go func() {
		defer close(c.done)
		value := 0
		for {
			select {
			case <-ctx.Done():
				return
			case req := <-c.requests:
				if req.ctx.Err() != nil {
					continue
				}
				value += req.delta
				// A caller can give up after handing over its request.
				// Never let its abandoned reply block the owner forever.
				select {
				case req.reply <- value:
				case <-req.ctx.Done():
				case <-ctx.Done():
					return
				}
			}
		}
	}()
	return c
}

// Add waits for both the request handoff and the owner's reply. Add(ctx, 0)
// reads a snapshot. The service context and the individual request may cancel.
func (c *Counter) Add(ctx context.Context, delta int) (int, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	req := request{ctx: ctx, delta: delta, reply: make(chan int)}
	select {
	case c.requests <- req:
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-c.done:
		return 0, fmt.Errorf("counter stopped")
	}
	select {
	case value := <-req.reply:
		return value, nil
	case <-ctx.Done():
		return 0, ctx.Err()
	case <-c.done:
		return 0, fmt.Errorf("counter stopped")
	}
}

func main() {
	fmt.Println("=== Unbuffered Channels: One Owner for Shared State ===")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	counter := NewCounter(ctx)
	var clients sync.WaitGroup
	for id := 1; id <= 3; id++ {
		clients.Add(1)
		go func() {
			defer clients.Done()
			for range 4 {
				value, err := counter.Add(ctx, 1)
				if err != nil {
					fmt.Printf("Client %d: %v\n", id, err)
					return
				}
				fmt.Printf("Client %d: owner replied with %d\n", id, value)
			}
		}()
	}
	clients.Wait()
	value, err := counter.Add(ctx, 0)
	fmt.Printf("Final counter: %d (expected 12), error: %v\n", value, err)
	canceled, stopRequest := context.WithCancel(ctx)
	stopRequest()
	_, err = counter.Add(canceled, 100)
	fmt.Println("Canceled request:", err)
	cancel()
	<-counter.done // Join the owner instead of sleeping and hoping it exited.
	fmt.Println("Main: all clients and the owner have stopped")
}
