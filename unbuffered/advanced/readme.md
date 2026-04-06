# Advanced Unbuffered Channel Example

This demonstrates a more complex use case: multiple workers with bidirectional communication using unbuffered channels.

Note: This example intentionally avoids `select` to focus on unbuffered channel basics. It is not meant to be a practical design, but rather to illustrate the synchronization behavior of unbuffered channels in a more complex scenario.  Technically this is a "request-response" pattern implemented with channels, but it is essentially just function calls with extra steps. The real power of channels emerges when you need timeouts, cancellation, or coordination between multiple goroutines, which this example intentionally avoids for simplicity.



## Architecture

### Worker Pattern
Each worker has:
- **taskChan**: Receives tasks from main (unbuffered)
- **doneChan**: Sends results back to main (unbuffered)

### Synchronization Points

Every task involves **two synchronization points**:

1. **Sending the task**: `w.taskChan <- task`
   - Main blocks until worker receives
   - Worker blocks until main sends

2. **Receiving the result**: `<-w.doneChan`
   - Worker blocks until main receives
   - Main blocks until worker sends

## Key Behaviors

### Sequential Execution
Because channels are unbuffered, `SendTask()` is **completely synchronous**:
```go
result := worker1.SendTask("Calculate Pi")  // Main waits here until task completes
fmt.Println(result)                         // Only executes after task finishes
```

Main cannot send another task until the current one completes. This makes the flow predictable but potentially slower.

### Guaranteed Handoffs
- Worker won't start processing until main has sent the task
- Worker won't continue after finishing until main has received the result
- Main knows exactly when the worker received the task (when send completes)
- Main knows exactly when the worker finished (when receive completes)

### Clean Shutdown
The shutdown process uses the same synchronization:
1. Main sends "quit" signal (blocks until worker receives)
2. Worker sends confirmation (blocks until main receives)
3. Main knows worker has fully stopped

## Comparison with Buffered Channels

With buffered channels:
- Main could send multiple tasks without waiting
- Workers could complete tasks while main is doing other work
- Less synchronization, more concurrency, but less predictability

With unbuffered channels (this example):
- Main and worker are tightly synchronized
- Clear request-response pattern
- Easier to reason about ordering
- Lower concurrency, but stronger guarantees

## The Honest Truth: Is This Overkill?

**Yes, for this example!** This pattern is basically **function calls with extra steps**. The worker could just be a regular function and it would be simpler.

### Why This Feels Like a Function Call

In this example, we're using unbuffered channels in a simple request-response pattern. Each task:
1. Blocks while sending
2. Blocks while waiting for result
3. Processes sequentially

This is exactly what a function call does, but with more code!

### Where Unbuffered Channels Actually Shine

The advantages appear when you add complexity this example avoids:

#### 1. **Timeouts** (requires select)
```go
select {
case w.taskChan <- task:
    // sent successfully
case <-time.After(1 * time.Second):
    // worker is stuck, handle timeout
}
```
Can't do this with function calls.

#### 2. **Cancellation** (requires select)
```go
select {
case result := <-w.doneChan:
    return result
case <-ctx.Done():
    return errors.New("cancelled")
}
```

#### 3. **Multiple Producers/Consumers**
When you have multiple goroutines racing to send/receive, channels coordinate them. With function calls, you'd need explicit locking.

#### 4. **Decoupling**
The sender doesn't need to know who the receiver is. You can pass channels around, wire up pipelines. Functions require direct coupling.

#### 5. **Non-blocking Checks** (requires select)
```go
select {
case ch <- value:
    // sent
default:
    // receiver not ready, do something else
}
```

### The Constraint

This example **intentionally avoids select** to focus on unbuffered channel basics. But that removes most of the advantages! Without select, timeouts, or multiple goroutines competing, unbuffered channels in a simple request-response pattern are indeed overkill.

**The real power of channels emerges when you need:**
- Coordination between multiple concurrent goroutines
- Timeouts and cancellation (see the `select` examples in this repo)
- Non-blocking operations
- Pipeline patterns where data flows through multiple stages

For simple sequential task processing like this example, just use regular functions.

## Running

```bash
go run main.go
```

You'll see each task complete before the next one starts, demonstrating the synchronous nature of unbuffered channels.
