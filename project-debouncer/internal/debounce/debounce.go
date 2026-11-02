// Package debounce implements trailing-edge debouncing with a single goroutine.
package debounce

import (
	"context"
	"fmt"
	"time"
)

// Stream emits the latest value after quiet has elapsed without another input.
// Closing input flushes the pending value immediately. Cancellation discards it.
// Output is unbuffered: a slow consumer applies backpressure, while newer input
// can still replace a pending value. The caller owns input; Stream owns output.
func Stream[T any](ctx context.Context, input <-chan T, quiet time.Duration) (<-chan T, error) {
	if quiet <= 0 {
		return nil, fmt.Errorf("quiet period must be positive")
	}
	output := make(chan T)
	go func() {
		defer close(output)
		timer := time.NewTimer(quiet)
		timer.Stop()
		defer timer.Stop()
		var latest T
		var pending bool
		var tick <-chan time.Time
		var send chan<- T
		for {
			if ctx.Err() != nil {
				return
			}
			select {
			case <-ctx.Done():
				return
			case value, ok := <-input:
				if !ok {
					if pending {
						select {
						case output <- latest:
						case <-ctx.Done():
						}
					}
					return
				}
				latest, pending = value, true
				send = nil
				timer.Reset(quiet) // Go 1.24 timer channels cannot deliver stale ticks.
				tick = timer.C
			case <-tick:
				tick = nil
				send = output
			case send <- latest:
				pending = false
				send = nil
			}
		}
	}()
	return output, nil
}
