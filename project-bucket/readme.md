# project-bucket

A tiny, self-contained Go project that demonstrates a token-bucket (leaky-bucket style) rate limiter.

What it shows
- Burst allowance: up to 3 immediate events
- Steady rate: ~1 token every 5ms
- How a producer bursts (batches every 50ms) but the consumer is smoothed by the bucket

Run (Windows PowerShell)
```powershell
# from project root
cd project-bucket

# build or run directly
go run ./cmd/bucket-demo
```

Run (bash/WSL)
```bash
cd project-bucket
go run ./cmd/bucket-demo
```

How it works
- internal/ratelimit/bucket.go implements a simple token bucket with explicit burst and interval
- cmd/bucket-demo emits batches every 50ms for ~1s, with this distribution:
  - 60%: 1 event
  - 10%: 3 events
  - 10%: 5 events
  - 10%: 6 events
  - 10%: 8 events
- The consumer calls bucket.Wait(ctx) per event, which enforces the rate and allows short bursts up to 3

Notes
- This bucket pre-fills to its burst capacity so the very first batch can burst instantly.
- The demo logs timestamps, produced, and consumed counts so you can see smoothing in action.
