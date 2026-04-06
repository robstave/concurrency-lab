# Atomic Operations in Go

## Overview

This subdirectory contains examples demonstrating the use of the `sync/atomic` package for lock-free, thread-safe integer operations.

## Concepts

When multiple goroutines access and modify the same variable concurrently, a race condition can occur. The standard way to prevent this in Go is to use a `sync.Mutex`. However, for simple counters or state flags, acquiring and releasing a lock can be relatively expensive and cause contention.

The `sync/atomic` package provides low-level atomic memory primitives useful for implementing synchronization algorithms. Atomic operations map directly to the corresponding CPU instructions, ensuring that a read-modify-write sequence on a variable happens entirely uninterrupted (atomically) and is instantly visible to other goroutines.

### Key functions:
- `atomic.AddInt64(&counter, 1)`: Safely increments an integer.
- `atomic.LoadInt64(&counter)`: Safely reads the value without acquiring a lock.
- `atomic.StoreInt64(&counter, val)`: Safely writes a value.
- `atomic.CompareAndSwapInt64(&val, old, new)`: Updates a value only if it matches an expected old value.

## Why use Atomic?
- **Performance**: Generally faster than using a `sync.Mutex` for simple counters.
- **Lock-free**: Avoids potential deadlocks associated with traditional mutexes.

## Running the Exercise

```bash
go run .
```
