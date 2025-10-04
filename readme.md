# concurrency-lab


Run the demo crawler with some sample targets and a rate limit of 5 req/s and 20 workers:


```bash
export URLS="https://example.com,https://httpbin.org/get,https://httpbin.org/bytes/102400,https://httpbin.org/delay/2"
go run ./cmd/crawler --concurrency 20 --rate 5 --timeout 5s --error-threshold 3


Then, open metrics at: http://localhost:2112/metrics

Flags (all optional):

--concurrency (default: 10)

--rate requests/second (default: 10)

--timeout per-request timeout (default: 4s)

--error-threshold cancel all work after this many worker errors (default: 10)

--port metrics server port (default: 2112)

Provide URLs via the URLS env var (comma-separated) or via STDIN, one per line.


echo -e "https://example.com\nhttps://httpbin.org/status/404" | go run ./cmd/crawler

