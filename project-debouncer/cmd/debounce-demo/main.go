package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"example.com/project-debouncer/internal/debounce"
)

func main() {
	quiet := flag.Duration("quiet", 150*time.Millisecond, "quiet period before emitting the latest event")
	flag.Parse()
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	input := make(chan string)
	output, err := debounce.Stream(ctx, input, *quiet)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		defer close(input)
		for _, word := range []string{"g", "go", "gor", "goro", "goroutine"} {
			select {
			case input <- word:
				fmt.Println("input:", word)
			case <-ctx.Done():
				return
			}
			timer := time.NewTimer(40 * time.Millisecond)
			select {
			case <-timer.C:
			case <-ctx.Done():
				timer.Stop()
				return
			}
		}
		// Keep input open long enough to demonstrate a quiet-period emission.
		timer := time.NewTimer(2 * *quiet)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-ctx.Done():
		}
	}()
	for value := range output {
		fmt.Println("debounced:", value)
	}
}
