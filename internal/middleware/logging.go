package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseRecorder struct {
	http.ResponseWriter
	status int
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

// Logging prints inbound request metadata, correlation ids and latency.
func Logging(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			recorder := &responseRecorder{
				ResponseWriter: writer,
				status:         http.StatusOK,
			}

			start := time.Now()
			next.ServeHTTP(recorder, request)

			logger.Info(
				"http request completed",
				"request_id", RequestIDFromContext(request.Context()),
				"user_id", UserIDFromContext(request.Context()),
				"method", request.Method,
				"path", request.URL.Path,
				"query", request.URL.RawQuery,
				"status_code", recorder.status,
				"duration_ms", time.Since(start).Milliseconds(),
				"user_agent", request.UserAgent(),
			)
		})
	}
}
