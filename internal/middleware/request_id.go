package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

// RequestID injects a correlation id into every request and response.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Прокси может уже проставить X-Request-ID — сохраняем для сквозной трассировки.
		requestID := request.Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
		}

		ctx := context.WithValue(request.Context(), requestIDKey, requestID)
		writer.Header().Set("X-Request-ID", requestID)

		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
