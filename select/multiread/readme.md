Read from multiple channels with select
======================================

This example shows how to read from two channels using a single `select` loop,
with a timeout fallback. It demonstrates that `select` chooses a ready case
non-deterministically when multiple cases can proceed.

What it does
------------
- Starts two goroutines that each send a message to `ch1` and `ch2` after a delay.
- The main loop runs two iterations and on each iteration does a `select` over:
	- receive from `ch1`
	- receive from `ch2`
	- timeout after 3 seconds (`time.After(3 * time.Second)`).
- After two successful receives (total), it signals `done` and exits.

Expected behavior (with current timings: 1s for ch1, 2s for ch2)
-----------------------------------------------------------------
- You will typically see something like:

	- received: from ch1
	- received: from ch2

- You should not see a timeout in this configuration because both messages arrive
	within 3 seconds. However, there is no guarantee in which order the messages
	are printed; `select` is not ordered, and scheduling can vary.

Non-determinism and no guarantees
---------------------------------
- If both channels are ready at the same time, `select` may pick either one.
- With two total messages, the loop ends after it has handled two receives, so
	you generally will not see the timeout case fire in this setup—but again,
	there’s no strict guarantee across runs or environments.

Timeout behavior by extending delays
------------------------------------
- If you change the sender goroutines to sleep for 10s and 20s instead of 1s and 2s:
	- The select loop has a 3s timeout. It will likely hit the timeout twice
		(two iterations) printing "timeout" each time.
	- After the two iterations, the loop completes and the `done` channel is sent,
		letting main exit. The late messages to `ch1`/`ch2` will be ignored since the
		loop is already finished.

Run it
------
From the repo root or this folder:

```
go run ./select/multiread
```

Try adjusting the delays in `main.go` to observe different outcomes.


https://go.dev/play/p/DWODRDZkYkf