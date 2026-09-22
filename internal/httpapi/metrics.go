package httpapi

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	requestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shortener_http_requests_total",
			Help: "Total HTTP requests by route and status.",
		}, []string{"route", "method", "status"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "shortener_http_request_duration_seconds",
			Help:    "HTTP request latency by route.",
			Buckets: prometheus.DefBuckets,
		}, []string{"route"},
	)

	clicksAttempted = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "shortener_clicks_attempted_total",
			Help: "Clicks handed to the recorder.",
		},
	)

	clicksDropped = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "shortener_clicks_dropped_total",
			Help: "Clicks discarded because the recorder queue was full.",
		},
	)
)

// ClicksDropped returns a callback that increments the dropped-click counter.
func ClicksDropped() func() {
	return clicksDropped.Inc
}

// statusRecorder captures the status code written by handler.
type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// withMetrics records request count and larency for a handler. route is a fixed label, never raw path, to keep cardinality bounded.
func withMetrics(route string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

		next(rec, r)

		requestsTotal.WithLabelValues(route, r.Method, strconv.Itoa(rec.status)).Inc()
		requestDuration.WithLabelValues(route).Observe(time.Since(start).Seconds())
	}
}
