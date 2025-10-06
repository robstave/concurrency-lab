package main

import (
	"bufio"
	"context"
	"flag"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"example.com/concurrency-lab/internal/logx"
	"example.com/concurrency-lab/internal/metrics"
	"example.com/concurrency-lab/internal/pipeline"
	"example.com/concurrency-lab/internal/ratelimit"
)

func main() {
	var (
		concurrency = flag.Int("concurrency", 10, "worker pool size")
		rate        = flag.Int("rate", 10, "requests per second")
		timeout     = flag.Duration("timeout", 4*time.Second, "per-request timeout")
		errThresh   = flag.Int("error-threshold", 10, "cancel after this many errors")
		port        = flag.Int("port", 2112, "metrics port")
	)
	flag.Parse()

	// Root context canceled on SIGINT/SIGTERM
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	log := logx.New()
	log.Infow("starting crawler", "concurrency", *concurrency, "rate", *rate, "timeout", timeout, "error_threshold", *errThresh, "port", *port)

	// Start metrics HTTP server (prometheus)
	stopMetrics := metrics.ServeAsync(*port)
	defer stopMetrics()

	// Prepare inputs: env var `URLS` or stdin
	urls := collectURLs()
	if len(urls) == 0 {
		log.Warnw("no URLs provided; nothing to do")
		return
	}

	// HTTP client with sane defaults
	client := &http.Client{Timeout: *timeout + time.Second}

	// Rate limiter shared across workers
	lim := ratelimit.NewTokenBucket(*rate, time.Second)
	defer lim.Stop()

	// Set up pipeline
	jobs := make(chan pipeline.Job, len(urls)) // buffer == total jobs; backpressure applies
	for _, u := range urls {
		jobs <- pipeline.Job{URL: u}
	}
	close(jobs)

	cfg := pipeline.Config{
		Concurrency:   *concurrency,
		PerRequestTO:  *timeout,
		ErrorCancelAt: *errThresh,
	}

	results, err := pipeline.Run(ctx, jobs, client, lim, cfg)
	if err != nil {
		log.Errorw("pipeline setup error", "error", err)
		os.Exit(1)
	}

	// Drain results until the workers finish or context is canceled
	for r := range results {
		if r.Err != nil {
			log.Infow("done", "url", r.URL, "status", r.Status, "bytes", r.Bytes, "ms", r.Latency.Milliseconds(), "err", r.Err)
			continue
		}
		log.Infow("done", "url", r.URL, "status", r.Status, "bytes", r.Bytes, "ms", r.Latency.Milliseconds())
	}

	log.Infow("crawler finished")
}

func collectURLs() []string {
	var urls []string

	// 1) Read from environment variable URLS (comma-separated list).
	// Example (PowerShell): $env:URLS="https://a.com,https://b.com"
	// Example (bash): export URLS="https://a.com,https://b.com"
	// We trim whitespace and then split on commas via splitCSV.
	if s := strings.TrimSpace(os.Getenv("URLS")); s != "" {
		urls = append(urls, splitCSV(s)...)
	}

	// 2) Optionally read from STDIN if input is piped.
	// We detect piped input by checking whether STDIN is a character device
	// (interactive terminal) or not. When you pipe data (e.g., `Get-Content urls.txt | go run ...`)
	// STDIN is NOT a terminal, so we proceed to read lines.
	stat, _ := os.Stdin.Stat()
	if (stat.Mode() & os.ModeCharDevice) == 0 { // piped input detected
		// Read one URL per line. Empty lines and surrounding whitespace are ignored.
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line != "" {
				urls = append(urls, line)
			}
		}
		// Note: scanner.Err() is intentionally ignored for simplicity; in a real
		// tool you might log or handle a read error here.
	}

	// Order: env var URLs first, then any piped URLs. No deduplication is performed.
	return urls
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
