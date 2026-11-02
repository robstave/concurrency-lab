package main

import (
	"context"
	"fmt"
	"sync"
	"time"
)

func produce(ctx context.Context, wg *sync.WaitGroup, count int) <-chan int {
	out := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out) // A stage closes only the channel it sends on.
		for n := 1; n <= count; n++ {
			fmt.Printf("Producer: offering %d\n", n)
			select {
			case out <- n:
				fmt.Printf("Producer: handed off %d\n", n)
			case <-ctx.Done():
				fmt.Println("Producer: canceled")
				return
			}
		}
	}()
	return out
}

func square(ctx context.Context, wg *sync.WaitGroup, input <-chan int) <-chan int {
	out := make(chan int)
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer close(out)
		for {
			select {
			case <-ctx.Done():
				return
			case n, ok := <-input:
				if !ok {
					return
				}
				fmt.Printf("Square: offering %d\n", n*n)
				select {
				case out <- n * n:
				case <-ctx.Done():
					fmt.Println("Square: canceled while sending")
					return
				}
			}
		}
	}()
	return out
}

func run(count, take int) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	var stages sync.WaitGroup
	output := square(ctx, &stages, produce(ctx, &stages, count))
	received := 0
	for value := range output {
		received++
		fmt.Printf("Consumer: received %d; spending 100ms processing it\n", value)
		time.Sleep(100 * time.Millisecond) // Simulated work, not a shutdown mechanism.
		if received == take {
			fmt.Println("Consumer: stopping early and canceling upstream")
			cancel()
			break
		}
	}
	stages.Wait() // Also safe when the consumer stopped without draining output.
	fmt.Printf("Main: received %d results; all stages have stopped\n", received)
}

func main() {
	fmt.Println("=== Unbuffered Pipeline: Slow Consumer and Backpressure ===")
	run(5, 0) // Consume everything; channel closure propagates downstream.
	fmt.Println("\n=== Early Exit: Cancel Instead of Abandoning Upstream ===")
	run(10, 3)
}
