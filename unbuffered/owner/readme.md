# Unbuffered Channels: One Owner for Shared State

## Concept

One goroutine owns a counter. Concurrent clients send increment requests over an
unbuffered channel and receive the new value on a private unbuffered reply channel.
Only the owner touches the counter, so the counter needs no mutex or atomic operations.

## Running

Requires Go 1.24+, consistent with the lab's other examples. From this directory:

```sh
go run main.go
go test -timeout 20s main.go main_test.go
```

Three clients each increment four times. The final value is **12**. Individual
client messages can appear out of numerical order because printing happens after
the reply handoff. Sending a delta of zero asks the owner for a snapshot.

## How It Works

1. A caller sends a request containing a delta, its context, and its reply channel.
2. The owner receives one request at a time and updates its private value.
3. The caller receives the reply before its `Add` call returns.
4. After clients finish, main cancels the service and waits for its `done` channel.

Both request and reply are handshakes. Accepting the request alone does not mean
the caller has received its answer. Each reply channel belongs to one request,
so concurrent callers cannot accidentally receive each other's results.

## Cancellation and Pitfalls

Both the request handoff and reply wait respect the caller's context. The owner
also checks that context when sending a reply: a caller abandoning its result
must not leave the owner stuck. Service cancellation stops the owner, and the
`done` channel signals that shutdown has actually completed. Clients never close
the shared requests channel; only the owner closes `done`.

A request canceled before processing is skipped. Cancellation after an update
does **not** roll it back: the caller may get a cancellation error even though the
counter changed. Retrying blindly can apply an update twice. This is an ownership
and coordination example, not a transaction or exactly-once protocol.

For a simple counter, a mutex or atomic operation is usually smaller and faster.
The ownership pattern becomes useful when a goroutine coordinates more complex
state transitions alongside other channel events.
