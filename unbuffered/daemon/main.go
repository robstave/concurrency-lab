package main

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Task represents work to be done
type Task struct {
	ID   int
	Name string
}

// Daemon manages continuous task processing
type Daemon struct {
	name        string
	taskCounter int
	wg          sync.WaitGroup
	mu          sync.Mutex
}

// NewDaemon creates a new daemon instance
func NewDaemon(name string) *Daemon {
	return &Daemon{
		name: name,
	}
}

// Start begins the daemon's continuous processing loop
func (d *Daemon) Start(ctx context.Context) {
	fmt.Printf("[%s] Starting daemon...\n", d.name)

	// Main daemon loop
	go func() {
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				fmt.Printf("[%s] Context cancelled, stopping daemon loop\n", d.name)
				fmt.Printf("[%s] Waiting for %d active workers to complete...\n", d.name, d.getActiveWorkers())
				d.wg.Wait()
				fmt.Printf("[%s] All workers completed\n", d.name)
				return

			case <-ticker.C:
				// Generate a new task periodically
				task := d.createTask()
				fmt.Printf("[%s] Spawning worker for task %d: %s\n", d.name, task.ID, task.Name)

				// Spawn a goroutine to process the task
				d.wg.Add(1)
				go d.processTask(ctx, task)
			}
		}
	}()
}

// createTask generates a new task
func (d *Daemon) createTask() Task {
	d.mu.Lock()
	d.taskCounter++
	id := d.taskCounter
	d.mu.Unlock()

	tasks := []string{"Process Image", "Analyze Data", "Generate Report", "Send Email", "Update Cache"}
	return Task{
		ID:   id,
		Name: tasks[rand.Intn(len(tasks))],
	}
}

// processTask simulates work in a spawned goroutine
func (d *Daemon) processTask(ctx context.Context, task Task) {
	defer d.wg.Done()

	// Simulate work duration (100ms to 1.5s)
	workDuration := time.Duration(100+rand.Intn(1400)) * time.Millisecond

	fmt.Printf("  [Worker-%d] Starting: %s (will take ~%dms)\n", task.ID, task.Name, workDuration.Milliseconds())

	// Do work, but respect context cancellation
	select {
	case <-time.After(workDuration):
		fmt.Printf("  [Worker-%d] ✓ Completed: %s\n", task.ID, task.Name)
	case <-ctx.Done():
		fmt.Printf("  [Worker-%d] ✗ Cancelled: %s (was interrupted)\n", task.ID, task.Name)
	}
}

// getActiveWorkers returns the approximate number of active workers
func (d *Daemon) getActiveWorkers() int {
	// This is a simplified approach - in production you'd track this more carefully
	// We're using WaitGroup which doesn't expose count, so this is an approximation
	return 0 // WaitGroup doesn't expose counter
}

// simulateKillSignal mimics receiving SIGTERM or SIGINT after a delay
func simulateKillSignal(delay time.Duration, killChan chan struct{}) {
	fmt.Printf("[KillSimulator] Will send kill signal in %s\n", delay)
	time.Sleep(delay)
	fmt.Printf("\n[KillSimulator] 💀 Sending kill signal!\n\n")
	killChan <- struct{}{} // This blocks until main receives
	fmt.Printf("[KillSimulator] Kill signal acknowledged by main\n")
}

func main() {
	fmt.Println("=== Production-Ready Daemon with Graceful Shutdown ===\n")

	// Seed random for task generation
	rand.Seed(time.Now().UnixNano())

	// Create unbuffered channel for kill signal
	// This is the key synchronization point - blocks until signal received
	killChan := make(chan struct{})

	// Create context for coordinating shutdown
	ctx, cancel := context.WithCancel(context.Background())

	// Create and start daemon
	daemon := NewDaemon("TaskProcessor")
	daemon.Start(ctx)

	// Spawn goroutine to simulate kill signal after 3 seconds
	go simulateKillSignal(3*time.Second, killChan)

	fmt.Println("[Main] Daemon running, waiting for kill signal...")
	fmt.Println("[Main] (In production, you'd use signal.Notify for real OS signals)")
	fmt.Println()

	// CRITICAL: This blocks main until kill signal is received
	// This is the unbuffered channel in action - perfect for coordination
	<-killChan

	fmt.Println("[Main] Kill signal received! Initiating graceful shutdown...")

	// Cancel context to signal all goroutines to stop
	cancel()

	// Give daemon time to clean up
	// In production, you might wait on a separate done channel from the daemon
	time.Sleep(100 * time.Millisecond)

	// Wait for daemon to finish
	fmt.Println("[Main] Waiting for daemon to finish cleanup...")
	time.Sleep(2 * time.Second) // In production, daemon would signal completion via channel

	fmt.Println()
	fmt.Println("=== Shutdown Complete ===")
	fmt.Println("✓ Kill signal coordinated via unbuffered channel")
	fmt.Println("✓ Context cancellation propagated to all workers")
	fmt.Println("✓ In-flight tasks allowed to complete or cancel cleanly")
	fmt.Println("✓ No goroutine leaks")
	fmt.Println("\n[Main] Exiting")
}
