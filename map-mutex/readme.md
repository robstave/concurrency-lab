# Protecting Maps with a Mutex

## Overview

This directory shows how to protect an in-memory shared map with a `sync.Mutex` or `sync.RWMutex` to prevent concurrent write races.

## Concepts

In Go, the built-in `map` data structure is **not safe for concurrent use**. If multiple goroutines attempt to read and write to the same map simultaneously, the Go runtime will detect a race condition and panic, crashing the application with `fatal error: concurrent map writes`.

### The Solution: sync.Mutex
To protect the map, we wrap it in a struct containing a `sync.Mutex` (or `sync.RWMutex`). Any goroutine that wants to interact with the map must first call `Lock()`, perform the operation, and then call `Unlock()`.

```go
type SafeMap struct {
    mu sync.Mutex
    m  map[string]int
}

func (sm *SafeMap) Set(key string, value int) {
    sm.mu.Lock()
    defer sm.mu.Unlock() // Use defer to guarantee unlocking even if a panic occurs
    sm.m[key] = value
}
```

### Mutex vs RWMutex
- `sync.Mutex`: Only one goroutine can hold the lock at a time. Period.
- `sync.RWMutex`: Allows *multiple* readers to hold a read lock (`RLock()`), as long as no writer holds the write lock (`Lock()`). This is significantly faster for read-heavy workloads.

*Note: Since Go 1.9, the standard library also offers `sync.Map` which is optimized for specific use cases, but for most general-purpose applications, a regular map heavily protected by an RWMutex is the idiomatic standard.*

## Running the Exercise

```bash
go run .
```
