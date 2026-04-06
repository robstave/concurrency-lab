package main

import (
	"context"
	"fmt"
	"time"
)

// Worker represents a task processor that respects context cancellation
type Worker struct {
	id       int
	taskChan chan string
	doneChan chan string
}

// NewWorker creates a new worker
func NewWorker(id int) *Worker {
	return &Worker{
		id:       id,
		taskChan: make(chan string), // unbuffered
		doneChan: make(chan string), // unbuffered
	}
}

// Start begins the worker's processing loop with context awareness
func (w *Worker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				fmt.Printf("Worker %d: Context cancelled, shutting down\n", w.id)
				return
			case task := <-w.taskChan:
				if task == "quit" {
					fmt.Printf("Worker %d: Received quit signal\n", w.id)
					w.doneChan <- "Worker shutdown"
					return
				}

				// Process the task with context awareness
				fmt.Printf("Worker %d: Processing '%s' (2 seconds)...\n", w.id, task)

				// Simulate work but check context periodically
				select {
				case <-time.After(2 * time.Second):
					// Work completed normally
					result := fmt.Sprintf("Completed: %s", task)

					// Try to send result, but respect context
					select {
					case w.doneChan <- result:
						fmt.Printf("Worker %d: Result sent successfully\n", w.id)
					case <-ctx.Done():
						fmt.Printf("Worker %d: Context cancelled while sending result, aborting\n", w.id)
						return
					}
				case <-ctx.Done():
					fmt.Printf("Worker %d: Context cancelled during processing, aborting\n", w.id)
					return
				}
			}
		}
	}()
}

// SendTaskWithTimeout sends a task with a timeout using context
func (w *Worker) SendTaskWithTimeout(task string, timeout time.Duration) (string, error) {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel() // Always clean up

	// Try to send task with context timeout
	select {
	case w.taskChan <- task:
		fmt.Println("Main: Task sent to worker")
	case <-ctx.Done():
		return "", fmt.Errorf("timeout sending task: %w", ctx.Err())
	}

	// Wait for result with context timeout
	select {
	case result := <-w.doneChan:
		return result, nil
	case <-ctx.Done():
		return "", fmt.Errorf("timeout waiting for result: %w", ctx.Err())
	}
}

func main() {
	fmt.Println("=== Unbuffered Channel with Context Cancellation ===\n")

	// Create a context for the entire application
	appCtx, appCancel := context.WithCancel(context.Background())
	defer appCancel()

	// Create and start worker
	worker := NewWorker(1)
	worker.Start(appCtx)

	// Give worker time to start
	time.Sleep(50 * time.Millisecond)

	// Example 1: Task that will timeout
	fmt.Println("--- Example 1: Timeout (1 second timeout, 2 second task) ---")
	result, err := worker.SendTaskWithTimeout("Process Large File", 1*time.Second)
	if err != nil {
		fmt.Printf("Main: ERROR - %s\n", err)
	} else {
		fmt.Printf("Main: Got result: %s\n", result)
	}

	// Give time to see worker's context cancellation message
	time.Sleep(200 * time.Millisecond)

	fmt.Println("\n--- Example 2: Successful task (3 second timeout, 2 second task) ---")

	// Recreate worker since previous one may have exited
	worker2 := NewWorker(2)
	worker2.Start(appCtx)
	time.Sleep(50 * time.Millisecond)

	result, err = worker2.SendTaskWithTimeout("Quick Task", 3*time.Second)
	if err != nil {
		fmt.Printf("Main: ERROR - %s\n", err)
	} else {
		fmt.Printf("Main: Got result: %s\n", result)
	}

	fmt.Println("\n--- Key Differences from 'cancel' example ---")
	fmt.Println("Worker checks context and exits cleanly (no goroutine leak)")
	fmt.Println("Worker won't block forever trying to send results")
	fmt.Println("Proper resource cleanup with defer cancel()")
	fmt.Println("Standard Go cancellation pattern using context.Context")

	// Cancel application context to clean up any remaining goroutines
	appCancel()
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\nMain: Exiting cleanly")
}
