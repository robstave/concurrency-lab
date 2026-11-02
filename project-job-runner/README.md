# Project: Background job runner

Run jobs through a bounded queue with a fixed worker pool, per-attempt deadlines,
limited retries, and cooperative shutdown. This is an in-memory teaching project.

Requires Go 1.24+. From this directory:

```sh
go run ./cmd/runner-demo
go run ./cmd/runner-demo -workers 2 -queue 2 -attempts 3 -timeout 2s -retry-delay 50ms
go test ./...
```

The demo has eight simulated jobs. Jobs 3 and 6 fail once and then succeed. Job 8
takes one second per attempt, so the default 300ms deadline exhausts its three
attempts and the demo deliberately exits with status 1. Use `-timeout 2s` for an
all-successful run. Results are printed after workers finish, in completion order.

`internal/runner.Run` accepts a job slice and configuration. The producer blocks
when the pending queue is full; at most `workers` tasks execute concurrently.
Queue capacity zero gives an unbuffered handoff. Each retry stays in its worker,
so attempts for one job never overlap. All task errors are retried up to `attempts`,
including attempt deadlines, with a fixed, cancellation-aware delay. Other jobs
continue even if a job exhausts its attempts.

Ctrl+C stops admission, cancels active tasks and retry waits, abandons queued jobs,
and waits for workers and producer to exit. Results cover jobs picked up by workers;
unstarted jobs have no result. Task failures are stored per result; configuration
errors and parent cancellation are returned as top-level errors. The CLI exits
nonzero if any task fails or the run is canceled.

Tasks must honor their context and be safe to retry. A deadline cannot forcibly
stop a Go function: an uncooperative task delays shutdown and occupies its worker.
Tasks must also join any goroutines they start and must not panic. This runner has
no persistence, restart recovery, or exactly-once guarantee. The job slice and final
results remain in memory even though the execution queue is bounded.
