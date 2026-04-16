package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "maxapp_http_requests_total",
			Help: "Total amount of handled HTTP requests.",
		},
		[]string{"method", "path", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "maxapp_http_request_duration_seconds",
			Help:    "Duration of handled HTTP requests.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

func init() {
	prometheus.MustRegister(httpRequestsTotal, httpRequestDuration)
}

// Metrics collects request counters and latency histograms.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Тот же приём, что в Logging: снимаем фактический статус после next.
		recorder := &responseRecorder{
			ResponseWriter: writer,
			status:         http.StatusOK,
		}

		start := time.Now()
		next.ServeHTTP(recorder, request)

		httpRequestsTotal.WithLabelValues(
			request.Method,
			request.URL.Path,
			strconv.Itoa(recorder.status),
		).Inc()
		httpRequestDuration.WithLabelValues(
			request.Method,
			request.URL.Path,
		).Observe(time.Since(start).Seconds())
	})
}
