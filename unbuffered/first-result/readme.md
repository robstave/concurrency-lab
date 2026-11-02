# Unbuffered Channels: First Successful Result Wins

## Concept

Race several simulated searches, accept the first successful response, and
cancel the others. A quick failure does not prevent a slower source from succeeding.

## Running

Requires Go 1.24+. From this directory:

```sh
go run main.go
go test -timeout 20s main.go main_test.go
```

Normally `fast-but-failing` reports an error after 20ms, `winning-search` succeeds
after 80ms, and `slow-search` is canceled before its 400ms search completes. These
are illustrative delays; scheduling determines which ready successful send wins.
The demo has a two-second overall deadline.

## How It Works

Every source has its own worker. They share one unbuffered results channel. Main
receives errors until a success arrives, all sources fail, or the parent context
expires. All failures are joined into one error; an empty source list returns an
error immediately.

On every return path, `FirstSuccess` cancels its worker context **before** waiting
for all workers. Each worker checks cancellation both during its simulated search
and when sending its result. The function returns only after every worker exits.

## The Abandoned Send Pitfall

After main accepts a winner, it stops receiving. A losing worker using a plain
`results <- result` could then block forever. The send must select between the
result handoff and cancellation. Canceling before `Wait` matters: waiting first
would deadlock on exactly those blocked senders.

No worker closes the shared results channel. The receiver counts up to one result
per source, and cancellation handles early exit, so closure is unnecessary. This
avoids one sender closing a channel while another is still sending.

This is a simulation, not a network search client. Real calls must propagate the
context into the operation itself; canceling a result send cannot forcibly stop
an uncooperative network call. If parent cancellation and a success become ready
together, either select case may win.



## To Run the Example

From this directory, execute:

```sh
go run main.go
go test -timeout 20s main.go main_test.go
```

## Results

```
=== Unbuffered Channels: First Successful Result Wins ===
Ignoring failed response: fast-but-failing unavailable
slow-search: search canceled
Winner: winning-search: found the answer
Main: all search workers have stopped
```
