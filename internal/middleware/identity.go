package middleware

import (
	"context"
	"net/http"
)

// Identity extracts a user id from headers so the scaffold can log user activity consistently.
func Identity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		userID := request.Header.Get("X-User-ID")
		if userID == "" {
			userID = request.URL.Query().Get("userId")
		}
		if userID == "" {
			userID = "100245"
		}

		ctx := context.WithValue(request.Context(), userIDKey, userID)
		next.ServeHTTP(writer, request.WithContext(ctx))
	})
}
