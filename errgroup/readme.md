Errgroup demo
==============

This example shows how to use golang.org/x/sync/errgroup to run several goroutines
and propagate the first error while cancelling the remaining work via context.

What it does
------------
- Starts three goroutines with an errgroup and a derived context.
- One goroutine returns an error after ~200ms.
- Another goroutine sleeps longer and would succeed if not cancelled.
- The third goroutine represents longer work and will be cancelled when the
  first error occurs.

Run
---
From the `errgroup/` folder:

```
go run ./cmd/eg-demo
```

Expected output (order may vary, but you should see the error printed and
worker3 cancelled):

```
worker2 failed
worker3 cancelled
errgroup finished with error: worker2 failed
```
