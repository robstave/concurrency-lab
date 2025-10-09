package main

import (
	"fmt"
	"sync"
)

// Worker function: reads integers from the jobs channel,
// processes them, and sends results to the results channel.
func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	for num := range jobs {
		// TODO: replace this with your processing logic.
		// Example: compute the square of the number.
		squared := num * num
		// Send the computed result back to the results channel.
		results <- squared
	}
}

func main() {
	// Example input slice. Replace with your data as needed.
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Create channels for jobs and results.
	jobs := make(chan int, len(numbers))
	results := make(chan int, len(numbers))

	// A WaitGroup to ensure we wait for all workers to finish.
	var wg sync.WaitGroup

	// Number of concurrent workers to run.
	workerCount := 3

	// Start the worker goroutines.
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker(i, jobs, results, &wg)
	}

	// Fan-out: send jobs to the workers.
	for _, n := range numbers {
		jobs <- n
	}
	close(jobs) // Close the jobs channel to signal no more data.

	// Wait for all workers to finish.
	go func() {
		wg.Wait()
		close(results) // Close results once all workers have finished.
	}()

	// Fan-in: collect the results and calculate the sum.
	sum := 0
	for result := range results {
		sum += result
	}

	fmt.Printf("Sum of squares: %d\n", sum)
}
