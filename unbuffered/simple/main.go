package main

import (
	"fmt"
	"time"
)

func main() {
	// Create an unbuffered channel
	ch := make(chan string)

	// Start a goroutine to send a message
	go func() {
		fmt.Println("Goroutine: About to send message...")
		ch <- "Hello from goroutine!" // This blocks until main receives
		fmt.Println("Goroutine: Message sent!")
	}()

	// Give the goroutine time to reach the send operation
	time.Sleep(100 * time.Millisecond)

	fmt.Println("Main: About to receive message...")
	msg := <-ch // This blocks until goroutine sends
	fmt.Println("Main: Received:", msg)

	// Give goroutine time to print its final message
	time.Sleep(100 * time.Millisecond)
}
