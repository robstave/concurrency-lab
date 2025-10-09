package main

import (
	"fmt"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
)

func main() {
	const N = 100

	// shared, concurrent-safe counts: key -> *int64
	var counts sync.Map

	// job queue
	jobs := make(chan int)

	// worker pool
	workerCount := runtime.NumCPU() * 2
	var wg sync.WaitGroup
	wg.Add(workerCount)

	for w := 0; w < workerCount; w++ {
		go func() {
			defer wg.Done()
			for n := range jobs {
				key := fizzBuzzKey(n)
				increment(&counts, key)
			}
		}()
	}

	// feed the jobs
	for i := 1; i <= N; i++ {
		jobs <- i
	}
	close(jobs)

	// wait for all
	wg.Wait()

	// gather keys to print in a nice order:
	// numeric keys sorted numerically, then Fizz/Buzz/Fizz Buzz
	var numericKeys []int
	var wordKeys []string

	counts.Range(func(k, v any) bool {
		s := k.(string)
		if isNumberKey(s) {
			n, _ := strconv.Atoi(s)
			numericKeys = append(numericKeys, n)
		} else {
			wordKeys = append(wordKeys, s)
		}
		return true
	})
	sort.Ints(numericKeys)
	sort.Strings(wordKeys) // Fizz, Fizz Buzz, Buzz (alphabetical)

	// print results
	fmt.Println("== counts by key ==")
	for _, n := range numericKeys {
		key := strconv.Itoa(n)
		fmt.Printf("%-9s -> %d\n", key, loadCount(&counts, key))
	}
	for _, key := range wordKeys {
		fmt.Printf("%-9s -> %d\n", key, loadCount(&counts, key))
	}

	// sanity check: total should be N
	var total int64
	counts.Range(func(_, v any) bool {
		total += atomic.LoadInt64(v.(*int64))
		return true
	})
	fmt.Printf("\nTotal processed: %d (expected %d)\n", total, N)
}

func fizzBuzzKey(n int) string {
	var b strings.Builder
	if n%3 == 0 {
		b.WriteString("Fizz")
	}
	if n%5 == 0 {
		if b.Len() > 0 {
			b.WriteByte(' ')
		}
		b.WriteString("Buzz")
	}
	if b.Len() == 0 {
		return strconv.Itoa(n)
	}
	return b.String()
}

func isNumberKey(s string) bool {
	_, err := strconv.Atoi(s)
	return err == nil
}

func increment(m *sync.Map, key string) {
	// store pointer to counter so multiple goroutines can atomically add
	actual, _ := m.LoadOrStore(key, new(int64))
	atomic.AddInt64(actual.(*int64), 1)
}

func loadCount(m *sync.Map, key string) int64 {
	if v, ok := m.Load(key); ok {
		return atomic.LoadInt64(v.(*int64))
	}
	return 0
}
