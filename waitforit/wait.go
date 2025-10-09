package main

import (
	"fmt"
	"sync"
)

// worker simulates a task performed by a worker goroutine.
// It prints a message and marks the task as done in the WaitGroup.
func worker(id int, wg *sync.WaitGroup) {
	defer wg.Done() // Decrement the counter when the goroutine completes.
	fmt.Printf("do a thing %d\n", id)
}

func main() {

	// A WaitGroup to ensure we wait for all workers to finish.
	var wg sync.WaitGroup

	// Number of concurrent workers to run.
	workerCount := 3

	// Start the worker goroutines.
	for i := 0; i < workerCount; i++ {
		wg.Add(1) // Increment the counter for each worker.
		go worker(i, &wg)
	}

	// Wait for all workers to complete.
	wg.Wait()

	fmt.Printf("done")
}
