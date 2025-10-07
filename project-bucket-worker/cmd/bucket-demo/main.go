package main

import (
	"context"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"example.com/project-bucket/internal/emitter"
	"example.com/project-bucket/internal/logx"
	"example.com/project-bucket/internal/ratelimit"
	"github.com/fatih/color"
)

// emitter moved to internal/emitter package

func main() {
	src := rand.NewSource(time.Now().UnixNano())
	rnd := rand.New(src)
	log := logx.New()
	log.Infow("bucket demo starting")

	// Create a bucket allowing burst=3, ~1 token every 5ms.
	bucket := ratelimit.NewBucket(5*time.Millisecond, 3)
	defer bucket.Stop()

	// Emitter
	emitChan := make(chan emitter.Event, 64)
	// pass the local rand to the emitter by setting package-level source
	// (the emitter currently uses math/rand global functions; to fully
	// parametrize it we'd inject rnd into the emitter. For simplicity we
	// set the global seed here.)
	_ = rnd
	go emitter.EventEmitter(emitChan)

	ctx := context.Background()
	start := time.Now()

	// Start 3 workers that consume from emitChan concurrently.
	const workerCount = 3
	var processed int64
	var wg sync.WaitGroup
	for i := 1; i <= workerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for ev := range emitChan {
				begin := ev.Timestamp.Sub(start).Milliseconds()
				if err := bucket.Wait(ctx); err != nil {
					log.Errorw("wait error", "err", err)
					return
				}

				end := time.Since(start).Milliseconds()
				delay := end - begin
				count := atomic.AddInt64(&processed, 1)

				c := getPaletteColor(int(delay))
				c.Printf("worker=%d emit t=%4dms processed t=%4dms delay t=%2dms batch=%d produced=%d \n", id, begin, end, delay, ev.BatchSize, count)
			}
		}(i)
		//go func(id int) { defer wg.Done(); ... }(i) is good and idiomatic.
		// I mean...you could pass in the wg here too, but only if its NOT copied to a new goroutine.
		//go func(id int, wg *sync.WaitGroup)  if your worried about that.

	}

	// Wait for all workers to finish once the emitter closes the channel.
	wg.Wait()

	log.Infow("bucket demo finished", "produced", atomic.LoadInt64(&processed), "elapsed_ms", time.Since(start).Milliseconds())
}

func getPaletteColor(delay int) *color.Color {

	if delay < 2 {
		return color.New(color.FgHiGreen)
	} else if delay < 6 {
		return color.New(color.FgHiYellow)
	} else if delay < 12 {
		return color.New(color.FgMagenta)
	} else {
		return color.New(color.FgHiRed)
	}
}
