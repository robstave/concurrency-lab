Select basics: non-blocking cases
=================================

This example demonstrates basic `select` usage with a buffered channel and
the `default` case to make non-blocking receive/send operations.

What it shows
-------------
- A non-blocking receive: if the channel has no value ready, the `default`
	branch runs ("no message received yet").
- Sending then receiving from a buffered channel: once a value is placed into
	the buffer, a subsequent non-blocking receive can get it.
- A non-blocking send: attempting to send to a full buffered channel will
	take the `default` branch ("channel full, couldn’t send").

Walkthrough
-----------
1) `ch := make(chan string, 1)` creates a buffered channel with capacity 1.
2) First `select` tries to receive immediately; since no value is present,
	 it hits `default`.
3) We send `"hello"` into `ch`, then a second `select` receives it.
4) For `ch2`, we pre-fill the buffer with `"first"`, then a non-blocking send
	 fails because the buffer is full (takes `default`).
5) After reading from `ch2`, the buffer frees up and a non-blocking send of
	 `"second"` now succeeds.

Run it
------

From the repo root or this folder:

```
go run ./select/example
```

You should see output along these lines:

```
no message received yet
received: hello
channel full, couldn’t send
received: first
sent second
received: second
```


see also https://go.dev/play/p/cSKPrxc5UZ6


