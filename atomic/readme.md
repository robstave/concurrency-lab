Atomic examples
===============

This folder contains small examples using Go's sync/atomic package.

1) atomic-demo (exactly 1000 via CAS)
------------------------------------
Increments a shared counter to exactly 1000 using a load + compare-and-swap (CAS)
loop across multiple goroutines. CAS ensures we never overshoot 1000 even under
race conditions.

Run

```
go run ./cmd/atomic-demo
```

Expected

```
final count=1000 (expected 1000)
```

2) atomic-demo2 (typed atomic counter)
-------------------------------------
Uses the typed atomic counter (`atomic.Uint64`) with 50 goroutines, each doing
1000 increments.

Run

```
go run ./cmd/atomic-demo2
```

Expected

```
ops: 50000
```

3) atomic-simple (typed atomic + atomic.Value)
----------------------------------------------
- Increments a typed atomic counter to 5000 in a simple loop.
- Shows an `atomic.Value` pattern by treating stored structs as immutable
  snapshots (creating a new value each iteration) for safe publication.

Run

```
go run ./cmd/atomic-simple
```

Expected

```
ops: 5000
final: A=5000 B=-4000 (expected A=5000 B=-4000)
```

Notes
-----
- These examples are intentionally minimal. In real programs, you might prefer
  channels or mutexes depending on the problem you’re solving.
