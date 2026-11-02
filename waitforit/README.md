# Waiting for goroutines

Three workers print a message, then main prints `done`. A `sync.WaitGroup` keeps main from returning before the workers finish.

Call `wg.Add(1)` before starting each worker. Each worker calls `defer wg.Done()`, and main calls `wg.Wait()` to block until the counter reaches zero.

Pass the WaitGroup by pointer when sharing it with a function. Copying it gives the worker a different counter.

Run from this directory:

```bash
go run wait.go


```



## Diagram

```mermaid
sequenceDiagram
    participant M as main
    participant WG as WaitGroup
    participant W0 as worker 0
    participant W1 as worker 1
    participant W2 as worker 2

    M->>WG: Add(1)
    M->>W0: go worker(0, &wg)
    M->>WG: Add(1)
    M->>W1: go worker(1, &wg)
    M->>WG: Add(1)
    M->>W2: go worker(2, &wg)
    M->>WG: Wait() (blocks)

    par concurrent, order not guaranteed
        W0->>W0: Printf("do a thing 0")
        W0->>WG: Done()
    and
        W1->>W1: Printf("do a thing 1")
        W1->>WG: Done()
    and
        W2->>W2: Printf("do a thing 2")
        W2->>WG: Done()
    end

    WG-->>M: counter == 0, unblock
    M->>M: Printf("done")
```