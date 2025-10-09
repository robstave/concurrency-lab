package main

import (
	"fmt"
)

func main() {
	// Example input slice. Replace with your data as needed.
	numbers := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Create a buffered channel to hold results.
	results := make(chan int, len(numbers))

	// Fan-out: send jobs to the workers.
	for _, n := range numbers {
		results <- n // Send each number to the results channel.
	}

	// Close the results channel to signal no more data.
	// If you comment out the close below, the program will deadlock
	// because the range loop will wait indefinitely for more data.
	close(results)

	// Accumulate the sum of numbers received from the channel.
	sum := 0
	for result := range results {
		fmt.Printf("Tally: %d\n", sum)
		sum += result
	}

	// Print the final sum of all numbers.
	fmt.Printf("Sum of numbers: %d\n", sum)
}
