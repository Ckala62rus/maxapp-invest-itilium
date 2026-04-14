package middleware

import (
	"net/http"
	"strings"
)

var allowedCORSOrigins = map[string]struct{}{
	"http://localhost:5173": {},
	"http://127.0.0.1:5173": {},
	"http://localhost:5174": {},
	"http://127.0.0.1:5174": {},
}

// CORS allows the local frontend dev server to call the API directly.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		origin := request.Header.Get("Origin")
		if origin == "" {
			next.ServeHTTP(writer, request)
			return
		}

		if !isAllowedCORSOrigin(origin) {
			if request.Method == http.MethodOptions {
				writer.WriteHeader(http.StatusForbidden)
				return
			}

			next.ServeHTTP(writer, request)
			return
		}

		headers := writer.Header()
		headers.Set("Access-Control-Allow-Origin", origin)
		headers.Set("Vary", "Origin")
		headers.Add("Vary", "Access-Control-Request-Method")
		headers.Add("Vary", "Access-Control-Request-Headers")
		headers.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		headers.Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-User-ID")

		if request.Method == http.MethodOptions {
			writer.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func isAllowedCORSOrigin(origin string) bool {
	if _, ok := allowedCORSOrigins[origin]; ok {
		return true
	}

	return strings.HasPrefix(origin, "http://localhost:517") || strings.HasPrefix(origin, "http://127.0.0.1:517")
}
