package main

import (
	"fmt"
	"time"
)

// Worker represents a task processor
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

// Start begins the worker's processing loop
func (w *Worker) Start() {
	go func() {
		for {
			// Wait to receive a task (blocks here)
			task := <-w.taskChan

			if task == "quit" {
				fmt.Printf("Worker %d: Shutting down\n", w.id)
				w.doneChan <- "Worker shutdown"
				return
			}

			// Process the task - takes 2 seconds!
			fmt.Printf("Worker %d: Processing '%s' (this will take 2 seconds)...\n", w.id, task)
			time.Sleep(2 * time.Second) // simulate long work

			// Send result back (blocks until main receives)
			result := fmt.Sprintf("Completed: %s", task)
			w.doneChan <- result
		}
	}()
}

// SendTask sends a task with a 1-second timeout
func (w *Worker) SendTask(task string) (string, error) {
	// Try to send task with timeout
	select {
	case w.taskChan <- task:
		// Task sent successfully, now wait for result
		fmt.Println("Main: Task sent to worker")
	case <-time.After(1 * time.Second):
		// Worker didn't receive the task within 1 second
		return "", fmt.Errorf("timeout: worker not ready to receive task")
	}

	// Wait for result with timeout
	select {
	case result := <-w.doneChan:
		// Got result
		return result, nil
	case <-time.After(1 * time.Second):
		// Worker didn't complete within 1 second
		return "", fmt.Errorf("timeout: worker did not complete task in time")
	}
}

// Shutdown signals the worker to stop and waits for confirmation
func (w *Worker) Shutdown() {
	w.taskChan <- "quit"
	<-w.doneChan // wait for shutdown confirmation
}

func main() {
	fmt.Println("=== Unbuffered Channel with Timeout Example ===\n")

	// Create and start worker
	worker := NewWorker(1)
	worker.Start()

	// Give worker time to start
	time.Sleep(50 * time.Millisecond)

	// Send a task - this will timeout!
	fmt.Println("Main: Sending task to worker...")
	fmt.Println("      (Worker takes 2 seconds, but we timeout after 1 second)")
	result, err := worker.SendTask("Process Large File")

	if err != nil {
		fmt.Printf("Main: ERROR - %s\n\n", err)
	} else {
		fmt.Printf("Main: Got result: %s\n\n", result)
	}

	// The worker is now stuck with a result it can't send!
	fmt.Println("Main: Note - worker is stuck trying to send result on doneChan")
	fmt.Println("      because we timed out and stopped listening")

	time.Sleep(500 * time.Millisecond)

	fmt.Println("\nMain: In a real system, you'd need proper cancellation")
	fmt.Println("      (context.Context, done channels, etc.)")
	fmt.Println("\nMain: Exiting (worker goroutine will be orphaned)")
}
