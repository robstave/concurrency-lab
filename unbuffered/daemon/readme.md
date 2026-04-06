# Production-Ready Daemon with Graceful Shutdown

This demonstrates a **realistic, production-like pattern** using unbuffered channels for coordinating daemon lifecycle and graceful shutdown.

## Architecture

```
Main
 ├─ Spawns: Daemon (continuous loop)
 │   └─ Spawns: Multiple Worker goroutines (short-lived tasks)
 └─ Spawns: Kill Signal Simulator (sends shutdown after delay)
 
Main blocks on: killChan (unbuffered)
Coordination via: context.Context
```

## The Pattern

### 1. Unbuffered Kill Channel
```go
killChan := make(chan struct{})
<-killChan  // Main blocks here until kill signal
```

**Why unbuffered?**
- Creates a synchronization point
- Guarantees that sender knows main received the signal
- Perfect for critical shutdown coordination
- Sender blocks until receiver is ready (ensures handshake)

### 2. Context Propagation
```go
ctx, cancel := context.WithCancel(context.Background())
daemon.Start(ctx)
// ... later ...
cancel()  // Signal all goroutines to stop
```

### 3. WaitGroup for Worker Tracking
```go
d.wg.Add(1)
go d.processTask(ctx, task)
// ... later ...
d.wg.Wait()  // Wait for all workers to finish
```

### 4. Graceful Shutdown Flow

1. **Kill signal arrives** → `killChan <- struct{}{}`
2. **Main unblocks** → Receives signal
3. **Cancel context** → `cancel()`
4. **Daemon stops spawning** → Receives `ctx.Done()`
5. **Daemon waits for workers** → `d.wg.Wait()`
6. **Workers complete or cancel** → Check `ctx.Done()` during work
7. **Daemon exits** → All workers finished
8. **Main exits** → Clean shutdown

## Key Features

### Continuous Task Generation
The daemon runs a ticker that spawns workers periodically:
```go
ticker := time.NewTicker(500 * time.Millisecond)
for {
    select {
    case <-ticker.C:
        go d.processTask(ctx, task)  // Spawn worker
    }
}
```

### Worker Lifecycle
Each worker:
- Does work (simulated with `time.Sleep`)
- Checks context to allow early cancellation
- Signals completion via WaitGroup

### Kill Signal Simulation
Instead of real OS signals (SIGTERM, SIGINT), we simulate with a goroutine:
```go
time.Sleep(3 * time.Second)
killChan <- struct{}{}  // Blocks until main receives
```

In production, you'd use `signal.Notify`:
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
<-sigChan  // Block until OS signal
```

## Why This is "Production-Ready"

✅ **Graceful shutdown** - Workers finish or cancel cleanly  
✅ **Context propagation** - Cancellation cascades properly  
✅ **WaitGroup tracking** - Know when workers are done  
✅ **No goroutine leaks** - Everything exits cleanly  
✅ **Unbuffered channel for critical coordination** - Ensures handshake  
✅ **Ticker for periodic work** - Common daemon pattern  
✅ **Randomized work** - More realistic than fixed delays  

## What Would Real Production Add?

This example is close, but real production needs:

### Logging
- Structured logging (zerolog, zap)
- Log levels and context
- Error tracking

### Metrics
- Active worker count
- Task completion rate
- Shutdown duration

### Error Handling
- Task failures and retries
- Panic recovery in workers
- Timeout enforcement

### Signal Handling
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
```

### Health Checks
- Liveness probe
- Readiness probe
- Graceful degradation

### Configuration
- Configurable timeouts
- Worker limits
- Shutdown deadlines

## The Unbuffered Channel's Role

The unbuffered `killChan` is perfect here because:

1. **Main must wait** - Can't exit until told to
2. **Sender confirms receipt** - Knows main got the signal
3. **Synchronization point** - Clear handoff moment
4. **Simple and clear** - No buffering complexity needed

A buffered channel would work, but doesn't provide the same guarantee that main received the signal.

## Common Patterns Demonstrated

### Daemon Pattern
- Continuous loop with ticker
- Spawns short-lived workers
- Coordinates many goroutines

### Graceful Shutdown Pattern
- Signal reception
- Context cancellation
- Worker draining
- Clean exit

### Coordination Pattern
- Unbuffered channel for critical sync
- Context for cancellation cascade
- WaitGroup for completion tracking

## Running

```bash
go run main.go
```

**What you'll see:**
1. Daemon starts spawning workers every 500ms
2. Workers process tasks (100ms-1.5s each)
3. After 3 seconds, kill signal arrives
4. Daemon stops spawning new tasks
5. In-flight workers complete or cancel
6. Clean shutdown

## Comparison with Earlier Examples

| Example | Purpose | Production-Ready? |
|---------|---------|-------------------|
| simple | Basic blocking | No - just learning |
| advanced | Worker pattern | No - overkill for what it does |
| cancel | Timeout pitfall | No - goroutine leaks |
| context | Proper cancellation | Closer - but still simplified |
| **daemon** | **Real-world pattern** | **Yes** - with caveats |

This daemon example is what you'd actually build in a real service.
