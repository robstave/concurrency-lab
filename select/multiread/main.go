package main

import (
	"fmt"
	"time"
)

func main() {
	ch1 := make(chan string)
	ch2 := make(chan string)
	done := make(chan bool)

	// Goroutine that sends on ch1 after a delay
	go func() {
		time.Sleep(1 * time.Second)
		ch1 <- "from ch1"
	}()

	// Goroutine that sends on ch2 after a longer delay
	go func() {
		time.Sleep(2 * time.Second)
		ch2 <- "from ch2"
	}()

	go func() {
		for range 2 {
			select {
			case msg1 := <-ch1: // read from ch1
				fmt.Println("received:", msg1)
			case msg2 := <-ch2: // read from ch2
				fmt.Println("received:", msg2)
			case <-time.After(3 * time.Second): // timeout
				fmt.Println("timeout")

			}
		}
		done <- true
	}()

	<-done

	fmt.Println("finished")
}
