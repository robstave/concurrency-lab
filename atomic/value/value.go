package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Config represents a snapshot of configuration shared between goroutines.
// Treat instances as immutable: when changing config, create a new value and
// publish it via atomic.Value. Readers see a consistent snapshot without locks.
type Config struct {
	Version        string
	Threshold      int
	FeatureEnabled bool
}

func main() {
	// Atomic holder for the latest config snapshot.
	var cfg atomic.Value

	// Initial config (version 1.0).
	cfg.Store(&Config{
		Version:        "1.0",
		Threshold:      10,
		FeatureEnabled: false,
	})

	// done is closed when the updater finishes its 3 updates.
	done := make(chan struct{})

	var wg sync.WaitGroup
	wg.Add(2)

	// 1) Updater: runs once a second, 3 times, then signals done.
	go func() {
		defer wg.Done()
		for i := 1; i <= 3; i++ {
			time.Sleep(1 * time.Second)
			cur := cfg.Load().(*Config)
			next := &Config{
				Version:        fmt.Sprintf("1.%d", i),
				Threshold:      cur.Threshold + 5,
				FeatureEnabled: !cur.FeatureEnabled,
			}
			cfg.Store(next) // Publish new snapshot atomically.
			fmt.Printf("updater: published config version=%s (threshold=%d feature=%v)\n",
				next.Version, next.Threshold, next.FeatureEnabled)
		}
		close(done) // Signal the reader to stop.
	}()

	// 2) Reader: every 500ms prints the currently active config version until done.
	go func() {
		defer wg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				snap := cfg.Load().(*Config) // Lock-free, consistent snapshot
				fmt.Printf("reader: current version=%s\n", snap.Version)
			case <-done:
				fmt.Println("reader: done signal received, exiting")
				return
			}
		}
	}()

	// Wait for both goroutines to finish.
	wg.Wait()

	// Show the final config snapshot.
	final := cfg.Load().(*Config)
	fmt.Printf("final config: version=%s threshold=%d feature=%v\n",
		final.Version, final.Threshold, final.FeatureEnabled)
}
