# Concurrent FizzBuzz (Overkill)

## Overview

This exercise features a concurrent worker-pool FizzBuzz implementation that aggregates results using a `sync.Map`. 

## Concepts

### The Exercise
FizzBuzz is traditionally a simple loop. Implementing it concurrently is usually "overkill," but serves as an excellent sandbox to learn advanced synchronization techniques without complicated business logic getting in the way.

### sync.Map
The standard Go `map` is not safe for concurrent use, requiring a `sync.RWMutex`. In Go 1.9, the `sync.Map` type was introduced. 
- `sync.Map` is optimized for two specific scenarios:
  1. When keys are written once but read many times (e.g., caches).
  2. When multiple goroutines read/write to *different* keys.
- Unlike a protected `map[string]int`, `sync.Map` requires type-assertion when loading/storing because it accepts `interface{}` values.
- Methods: `Load()`, `Store()`, `LoadOrStore()`, `Delete()`, and `Range()`.

### Bounded Parallelism
Here we instantiate a pool of workers reading from a shared channel of numbers. Once each goroutine computes the result, it is stored in the shared `sync.Map` avoiding race conditions.

## Running the Exercise

```bash
go run .
```
