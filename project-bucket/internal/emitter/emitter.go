package emitter

import (
	"math/rand"
	"time"
)

// Event represents a single emitted item with the source timestamp and batch size.
type Event struct {
	Timestamp time.Time
	BatchSize int
}

// EventEmitter emits batches every 50ms over 1s (20 ticks total).
// Batch sizes by probability (matches pickBatch):
// - 60%: 1 event
// - 10%: 3 events
// - 10%: 5 events
// - 10%: 6 events
// - 10%: 8 events
func EventEmitter(out chan<- Event) {
	defer close(out)
	end := time.Now().Add(1 * time.Second)
	t := time.NewTicker(50 * time.Millisecond)
	defer t.Stop()
	for time.Now().Before(end) {
		<-t.C
		batch := pickBatch()

		ev := Event{
			Timestamp: time.Now(),
			BatchSize: batch,
		}

		for i := 0; i < batch; i++ {
			out <- ev
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
