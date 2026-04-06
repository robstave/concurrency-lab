# Unbuffered Channel Examples

This directory contains a series of examples demonstrating unbuffered channels in Go, from basic concepts to real-world patterns and common pitfalls.

> **Note**: These examples are intentionally simplified to focus on specific concepts. They're not production-ready code, but rather teaching tools to understand how unbuffered channels work and where they shine (or don't).

## The Examples

### 1. [simple](simple/) - Basic Unbuffered Channel
**Concept**: Synchronization between goroutines

The simplest possible example: one goroutine sends, main goroutine receives. Shows the fundamental blocking behavior of unbuffered channels.

**Key Takeaway**: Unbuffered channels create synchronization points - both sender and receiver must be ready at the same time.

**Limitation**: No select, no timeouts, just pure blocking behavior.

---

### 2. [advanced](advanced/) - Multiple Workers Pattern
**Concept**: Request-response with multiple workers

Shows a more structured pattern with worker structs, bidirectional channels, and coordinated shutdown.

**Key Takeaway**: This demonstrates that simple request-response with unbuffered channels is **basically just function calls with extra steps**. Without select, timeouts, or multiple goroutines racing, channels add complexity without much benefit.

**Honest Assessment**: Overkill for this use case! But it sets up patterns needed for the next examples.

---

### 3. [cancel](cancel/) - Timeouts with Select (The Problem)
**Concept**: Using select for timeouts - and the pitfall

Shows what happens when you add timeouts using `select` and `time.After()`, but **don't handle cancellation properly**.

**Key Takeaway**: 
- ✅ Select enables timeouts (this is where channels shine vs function calls!)
- ❌ Without proper cancellation, you leak goroutines
- ❌ Worker gets stuck trying to send results nobody is listening for

**Why Show This?**: Because it's a **common mistake**. This example shows what NOT to do, which is valuable for learning.

---

### 4. [context](context/) - Proper Cancellation (The Solution)
**Concept**: Using context.Context for clean cancellation

The **correct way** to handle timeouts with unbuffered channels. Shows how `context.Context` allows workers to:
- Detect when they've been cancelled
- Exit cleanly without leaking goroutines
- Follow idiomatic Go patterns

**Key Takeaway**: This is the production-ready pattern. When you need timeouts with goroutines and channels, use context.

**Comparison**: Same timeout scenario as `cancel`, but no goroutine leaks.

---

### 5. [daemon](daemon/) - Production-Ready Daemon Pattern
**Concept**: Continuous processing with graceful shutdown

A **realistic, production-like pattern** showing a daemon that:
- Runs continuously, spawning worker goroutines
- Uses unbuffered channel to block main until kill signal
- Coordinates shutdown via context cancellation
- Tracks in-flight workers with sync.WaitGroup
- Demonstrates graceful shutdown flow

**Key Takeaway**: This is where unbuffered channels **really shine** - blocking main for coordination while daemon runs in background. Shows how all the pieces (unbuffered channels, context, WaitGroup) work together in a real service pattern.

**Production-Ready**: Yes (with caveats noted in readme) - this is the pattern you'd actually build.

---

## The Progression

These examples build on each other:

1. **simple**: Here's how unbuffered channels block
2. **advanced**: Here's a pattern with workers (but it's overkill without select)
3. **cancel**: Here's select with timeouts (powerful! but broken)
4. **context**: Here's the proper way to do it
5. **daemon**: Here's how it all comes together in a real service

## When to Actually Use Unbuffered Channels

Based on these examples, use unbuffered channels when:

### ✅ Good Use Cases
- You need guaranteed synchronization between goroutines
- **Blocking main until shutdown signal** (daemon pattern)
- You're implementing pipelines with backpressure
- You want to coordinate work across multiple producers/consumers
- You need timeouts or cancellation (with context!)
- You're using select to coordinate multiple channels
- Critical coordination points where handshake semantics matter

### ❌ Probably Overkill
- Simple sequential task processing (just use functions)
- Request-response without timeouts
- When buffered channels would be simpler
- Single producer, single consumer with no coordination needs

## Key Concepts Demonstrated

### Blocking Behavior
- Send blocks until receiver is ready
- Receive blocks until sender is ready
- This creates a "handshake" between goroutines

### Select Statement
- Enables timeouts (`time.After`)
- Enables cancellation (`<-ctx.Done()`)
- Enables non-blocking operations (`default`)
- **This is where channels become powerful**

### Context Cancellation
- Standard Go pattern for goroutine lifecycle management
- Prevents goroutine leaks
- Enables graceful shutdown
- Required for production code

### Common Pitfalls
- Goroutine leaks when using timeouts without cancellation
- Deadlocks when both sides are waiting
- Orphaned goroutines that can't exit

## Not Covered (But Important)

These examples intentionally omit some important topics:
- Buffered channels (different tradeoffs)
- Channel closing and range loops
- Fan-out/fan-in patterns (see other directories)
- Worker pools with sync.WaitGroup
- Error handling and recovery

## Running the Examples

Each subdirectory has its own `main.go` and `readme.md`. Navigate to any example and run:

```bash
cd simple        # or advanced, cancel, context, daemon
go run main.go
```

## The Honest Truth

Most of these examples are **teaching tools**, not production blueprints. However, the **daemon** example is close to production-ready and shows realistic patterns.

Real production systems need:
- Proper error handling
- Metrics and observability  
- Graceful shutdown mechanisms (✓ shown in daemon)
- Resource limits and backpressure
- Recovery from panics
- Testing and validation
- Real OS signal handling (not simulated)

But understanding these basic patterns is essential before building those real systems. The daemon example bridges the gap between teaching and reality.

## Recommended Reading Order

### For Learning the Basics
1. Start with [simple](simple/) to understand blocking
2. Look at [advanced](advanced/) and read the "Honest Truth" section
3. See [cancel](cancel/) to understand the timeout problem
4. Study [context](context/) to learn the proper solution
5. Finish with [daemon](daemon/) to see it all together in a realistic pattern

### For Building Real Services
If you just want to see production patterns, jump straight to [daemon](daemon/) - but you'll get more from it if you understand the progression.

This progression shows both **what works** and **what doesn't**, which is more valuable than just showing best practices. The journey from simple blocking to production daemon reveals why each piece matters.
