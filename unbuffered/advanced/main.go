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
		fmt.Printf("Worker %d: ready\n", w.id)
		for {
			// Wait to receive a task (blocks here)
			task := <-w.taskChan

			if task == "quit" {
				fmt.Printf("Worker %d: Shutting down\n", w.id)
				w.doneChan <- "Worker shutdown"
				return
			}

			// Process the task
			fmt.Printf("Worker %d: Processing '%s'\n", w.id, task)
			time.Sleep(100 * time.Millisecond) // simulate work

			// Send result back (blocks until main receives)
			result := fmt.Sprintf("Completed: %s", task)
			w.doneChan <- result
		}
	}()
}

// SendTask sends a task and waits for completion
func (w *Worker) SendTask(task string) string {
	// Send task (blocks until worker receives)
	w.taskChan <- task

	// Wait for result (blocks until worker sends)
	return <-w.doneChan
}

// Shutdown signals the worker to stop and waits for confirmation
func (w *Worker) Shutdown() {
	w.taskChan <- "quit"
	<-w.doneChan // wait for shutdown confirmation
}

func main() {
	fmt.Println("=== Unbuffered Channel: Multiple Workers Example ===\n")

	// Create workers
	worker1 := NewWorker(1)
	worker2 := NewWorker(2)

	// Start workers
	worker1.Start()
	worker2.Start()

	// Give workers time to start
	time.Sleep(50 * time.Millisecond)

	// Send tasks to worker 1
	fmt.Println("Main: Sending task to worker 1...")
	result := worker1.SendTask("Calculate Pi")
	fmt.Printf("Main: Got result: %s\n\n", result)

	// Send tasks to worker 2
	fmt.Println("Main: Sending task to worker 2...")
	result = worker2.SendTask("Read Database")
	fmt.Printf("Main: Got result: %s\n\n", result)

	// Send another task to worker 1
	fmt.Println("Main: Sending second task to worker 1...")
	result = worker1.SendTask("Process Image")
	fmt.Printf("Main: Got result: %s\n\n", result)

	// Demonstrate synchronous nature: main blocks during each task
	fmt.Println("Main: All tasks completed sequentially due to unbuffered channels")
	fmt.Println("      (each SendTask blocks until worker finishes)\n")

	// Shutdown workers
	fmt.Println("Main: Shutting down workers...")
	worker1.Shutdown()
	worker2.Shutdown()

	fmt.Println("Main: All workers shut down. Exiting.")
}
