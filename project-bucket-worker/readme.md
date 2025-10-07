# project-bucket-worker

A small Go demo showing a token-bucket rate limiter driving a pool of workers.

What it shows
- Burst allowance: up to 3 immediate events
- Steady rate: ~1 token every 5ms
- Three concurrent workers, each adding a 10ms processing delay per item
- Producer bursts (batches every 50ms), consumer side is smoothed by the bucket and parallelism

Run (Windows PowerShell)
```powershell
# from repository root
cd project-bucket-worker

# run
go run ./cmd/bucket-demo
```

Run (bash/WSL)
```bash
cd project-bucket-worker
go run ./cmd/bucket-demo
```

How it works
- `internal/ratelimit/bucket.go` implements a token bucket with explicit burst and refill interval
- `cmd/bucket-demo` emits batches every 50ms for ~1s, with this distribution:
  - 60%: 1 event
  - 10%: 3 events
  - 10%: 5 events
  - 10%: 6 events
  - 10%: 8 events
- Three workers read from the emitter channel concurrently; each item waits for a bucket token and then sleeps 10ms to simulate work

Notes
- The bucket pre-fills to its burst capacity so the first batch can burst instantly.
- Logs include worker id, timings, and running count so you can observe smoothing and concurrency.
