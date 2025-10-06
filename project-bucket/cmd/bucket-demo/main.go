package main

import (
	"context"
	"math/rand"
	"time"

	"example.com/project-bucket/internal/logx"
	"example.com/project-bucket/internal/ratelimit"
	"github.com/fatih/color"
)

// eventEmitter emits batches every 50ms over 1s (20 ticks total).
// Batch sizes by probability:
// - 60%: 1 event
// - 10%: 2 events
// - 10%: 3 events
// - 10%: 4 events
// - 10%: 5 events
func eventEmitter(out chan<- int) {
	defer close(out)
	end := time.Now().Add(1 * time.Second)
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()
	for time.Now().Before(end) {
		<-t.C
		batch := pickBatch()
		for range batch {
			out <- batch // just a placeholder payload
		}
	}
}

func pickBatch() int {
	x := rand.Float64()
	switch {
	case x < 0.60:
		return 1
	case x < 0.70:
		return 3
	case x < 0.80:
		return 5
	case x < 0.90:
		return 6
	default:
		return 8
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log := logx.New()
	log.Infow("bucket demo starting")

	// Create a bucket allowing burst=3, ~1 token every 5ms.
	bucket := ratelimit.NewBucket(5*time.Millisecond, 3)
	defer bucket.Stop()

	// Emitter
	emitChan := make(chan int, 64)
	go eventEmitter(emitChan)

	ctx := context.Background()
	start := time.Now()
	var produced int

	for batch := range emitChan {

		produced++

		begin := time.Since(start).Milliseconds()

		if err := bucket.Wait(ctx); err != nil {
			log.Errorw("wait error", "err", err)
			break
		}

		end := time.Since(start).Milliseconds()
		delay := end - begin

		c := getPaletteColor(int(delay))
		c.Printf("emit t=%4dms processd t=%4dms delay t=%2dms batch=%d produced=%d \n", begin, end, delay, batch, produced)
	}

	log.Infow("bucket demo finished", "produced", produced, "elapsed_ms", time.Since(start).Milliseconds())
}

func getPaletteColor(delay int) *color.Color {

	if delay < 1 {
		return color.New(color.FgHiGreen)
	} else if delay < 2 {
		return color.New(color.FgHiYellow)
	} else if delay < 4 {
		return color.New(color.FgMagenta)
	} else {
		return color.New(color.FgHiRed)
	}
}
