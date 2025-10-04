package metrics

import (
	"fmt"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	requests = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "crawler_requests_total",
		Help: "Total requests attempted",
	}, []string{"url"})

	success = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "crawler_success_total",
		Help: "Total successful responses (non-5xx)",
	}, []string{"url"})

	latency = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "crawler_latency_seconds",
		Help:    "Latency of HTTP requests",
		Buckets: prometheus.DefBuckets,
	}, []string{"url"})

	bytesRead = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Name:       "crawler_bytes_read",
		Help:       "Bytes read per response",
		Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.99: 0.001},
	}, []string{"url"})
)

func init() {
	prometheus.MustRegister(requests, success, latency, bytesRead)
}

func ObserveRequest(url string, status int, durSecondsFmt interface{}, n int64, ok bool) {
	requests.WithLabelValues(url).Inc()
	latency.WithLabelValues(url).Observe(toSeconds(durSecondsFmt))
	bytesRead.WithLabelValues(url).Observe(float64(n))
	if ok {
		success.WithLabelValues(url).Inc()
	}
}

func toSeconds(v interface{}) float64 {
	switch x := v.(type) {
	case int64:
		return float64(x) / 1e9
	case float64:
		return x
	case interface{ Seconds() float64 }:
		return x.Seconds()
	default:
		return 0
	}
}

// ServeAsync starts a metrics server and returns a stopper func.
func ServeAsync(port int) func() {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	server := &http.Server{Addr: fmt.Sprintf(":%d", port), Handler: mux}
	go func() { _ = server.ListenAndServe() }()
	return func() { _ = server.Close() }
}
