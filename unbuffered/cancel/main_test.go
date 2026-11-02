package main

import (
	"context"
	"testing"
	"time"
)

func assertReply(t *testing.T, result string, err error, wantResult, wantError string) {
	t.Helper()
	if result != wantResult {
		t.Errorf("result = %q; want %q", result, wantResult)
	}
	if wantError == "" {
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	} else if err == nil || err.Error() != wantError {
		t.Errorf("error = %v; want %q", err, wantError)
	}
}

// Use a watchdog rather than letting a broken shutdown hang the entire suite.
func shutdown(t *testing.T, worker *Worker) {
	t.Helper()
	done := make(chan struct{})
	go func() {
		worker.Shutdown()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(4 * time.Second):
		t.Fatal("Shutdown did not receive the worker's acknowledgment")
	}
}

func TestSendTaskTimesOutWithoutReceiver(t *testing.T) {
	t.Parallel()
	// Do not start a worker: nobody can accept the unbuffered request.
	worker := NewWorker(1)
	result, err := worker.SendTask("unaccepted task")
	assertReply(t, result, err, "", "timeout: worker not ready to receive task")
}

func TestSendTaskReturnsSuccessfulReply(t *testing.T) {
	t.Parallel()
	worker := NewWorker(2)
	ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
	done := make(chan struct{})
	received := make(chan string, 1)
	// The real worker always takes longer than SendTask's deadline. Use a
	// cooperative peer to exercise SendTask's successful request/reply path.
	go func() {
		defer close(done)
		select {
		case task := <-worker.taskChan:
			received <- task
		case <-ctx.Done():
			return
		}
		select {
		case worker.doneChan <- "Completed: quick task":
		case <-ctx.Done():
		}
	}()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(4 * time.Second):
			t.Error("reply peer did not exit")
		}
	})
	result, err := worker.SendTask("quick task")
	assertReply(t, result, err, "Completed: quick task", "")
	select {
	case task := <-received:
		if task != "quick task" {
			t.Errorf("request = %q; want quick task", task)
		}
	default:
		t.Error("request was not handed to the reply peer")
	}
}

func TestTimedOutTaskBlocksNextRequestUntilResultDrained(t *testing.T) {
	t.Parallel()
	worker := NewWorker(3)
	worker.Start()
	// Even if an assertion fails, drain the abandoned result before shutdown.
	// Otherwise this deliberately broken example would leak into other tests.
	t.Cleanup(func() {
		select {
		case result := <-worker.doneChan:
			if result != "Completed: slow task" {
				t.Errorf("late result = %q; want Completed: slow task", result)
			}
		case <-time.After(4 * time.Second):
			t.Error("timed-out task never produced its late result")
		}
		shutdown(t, worker)
	})
	result, err := worker.SendTask("slow task")
	assertReply(t, result, err, "", "timeout: worker did not complete task in time")

	// SendTask gave up, but Start did not cancel its two-second task. Once
	// finished, it is stuck sending the first result rather than reading jobs.
	result, err = worker.SendTask("next task")
	assertReply(t, result, err, "", "timeout: worker not ready to receive task")
}

func TestShutdownIdleWorker(t *testing.T) {
	t.Parallel()
	worker := NewWorker(4)
	worker.Start()
	shutdown(t, worker)
}

func TestShutdownWaitsForAbandonedResult(t *testing.T) {
	t.Parallel()
	worker := NewWorker(5)
	worker.Start()
	// Bypass SendTask's timeout to hand the worker a job directly, and leave
	// its result unread. Shutdown cannot be accepted until that send finishes.
	select {
	case worker.taskChan <- "abandoned task":
	case <-time.After(4 * time.Second):
		t.Fatal("worker did not accept the task")
	}
	done := make(chan struct{})
	go func() {
		worker.Shutdown()
		close(done)
	}()
	// Always release the pending result and wait for shutdown, even if the
	// assertion below fails. No demonstration goroutine is left behind.
	t.Cleanup(func() {
		select {
		case result := <-worker.doneChan:
			if result != "Completed: abandoned task" {
				t.Errorf("late result = %q; want Completed: abandoned task", result)
			}
		case <-time.After(4 * time.Second):
			t.Error("worker did not deliver its abandoned result")
		}
		select {
		case <-done:
		case <-time.After(4 * time.Second):
			t.Error("shutdown remained blocked after draining the result")
		}
	})
	select {
	case <-done:
		t.Error("Shutdown returned while the worker still had an unread result")
	case <-time.After(2100 * time.Millisecond):
		// Wait beyond the two-second work duration without draining the result.
		// Finishing the work alone cannot release the shutdown handshake.
	}
}
