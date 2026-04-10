package middleware

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/models"
)

// Recover shields the API from panics and returns a structured response.
func Recover(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			defer func() {
				if recovered := recover(); recovered != nil {
					logger.Error(
						"http panic recovered",
						"request_id", RequestIDFromContext(request.Context()),
						"user_id", UserIDFromContext(request.Context()),
						"path", request.URL.Path,
						"panic", recovered,
					)

					writeJSON(writer, http.StatusInternalServerError, models.APIResponse{
						Success:   false,
						Message:   "internal server error",
						RequestID: RequestIDFromContext(request.Context()),
					})
				}
			}()

			next.ServeHTTP(writer, request)
		})
	}
}

// writeJSON keeps middleware self-sufficient for panic handling.
func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}
