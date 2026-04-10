// Package middleware contains reusable HTTP middleware.
package middleware

import "context"

type contextKey string

const (
	requestIDKey contextKey = "request_id"
	userIDKey    contextKey = "user_id"
)

// RequestIDFromContext returns the current request id from context when it exists.
func RequestIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(requestIDKey).(string)
	return value
}

// UserIDFromContext returns the current user id from context when it exists.
func UserIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(userIDKey).(string)
	return value
}
