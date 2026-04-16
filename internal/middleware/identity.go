package middleware

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

// AccessTokenVerifier validates backend access tokens and returns a trusted user id.
type AccessTokenVerifier interface {
	// ParseAccessToken validates the backend bearer token and returns trusted claims.
	ParseAccessToken(token string, now time.Time) (struct{ UserID string }, error)
}

// Identity resolves the current user id from a backend access token or explicit debug headers.
func Identity(logger *slog.Logger, verifier AccessTokenVerifier, allowDebugHeaders bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			ctx := request.Context()
			token := bearerTokenFromHeader(request.Header.Get("Authorization"))
			// 1) Приоритет: валидный Bearer — доверяем userId из нашего access-токена после MAX-валидации.
			if token != "" && verifier != nil {
				claims, err := verifier.ParseAccessToken(token, time.Now().UTC())
				if err == nil && strings.TrimSpace(claims.UserID) != "" {
					if logger != nil {
						logger.Info(
							"request identity resolved",
							"source", "access_token",
							"user_id", strings.TrimSpace(claims.UserID),
							"path", request.URL.Path,
						)
					}
					ctx = context.WithValue(ctx, userIDKey, strings.TrimSpace(claims.UserID))
					next.ServeHTTP(writer, request.WithContext(ctx))
					return
				}
				if logger != nil {
					logger.Warn(
						"request identity token rejected",
						"path", request.URL.Path,
						"error", err,
					)
				}
			}

			// 2) Только для локальной отладки: без токена можно передать X-User-ID или ?userId= (если включено в конфиге).
			if allowDebugHeaders && token == "" {
				userID := strings.TrimSpace(request.Header.Get("X-User-ID"))
				if userID == "" {
					userID = strings.TrimSpace(request.URL.Query().Get("userId"))
				}
				if userID != "" {
					if logger != nil {
						logger.Warn(
							"request identity resolved",
							"source", "debug_header",
							"user_id", userID,
							"path", request.URL.Path,
						)
					}
					ctx = context.WithValue(ctx, userIDKey, userID)
				}
			}

			// Пустой userId в контексте — норма для публичных маршрутов; защищённые отсекает RequireIdentity.
			if logger != nil && strings.TrimSpace(UserIDFromContext(ctx)) == "" {
				logger.Debug(
					"request identity is empty",
					"path", request.URL.Path,
					"allow_debug_headers", allowDebugHeaders,
					"has_bearer_token", token != "",
				)
			}

			next.ServeHTTP(writer, request.WithContext(ctx))
		})
	}
}

// RequireIdentity blocks protected routes when no trusted user id is present.
func RequireIdentity(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// После Identity userId должен быть в контексте; иначе клиент не прошёл ни токен, ни debug-режим.
		if strings.TrimSpace(UserIDFromContext(request.Context())) == "" {
			writeJSON(writer, http.StatusUnauthorized, map[string]any{
				"success":   false,
				"message":   "authentication is required",
				"requestId": RequestIDFromContext(request.Context()),
			})
			return
		}

		next.ServeHTTP(writer, request)
	})
}

func bearerTokenFromHeader(headerValue string) string {
	headerValue = strings.TrimSpace(headerValue)
	if headerValue == "" {
		return ""
	}

	const prefix = "Bearer "
	if !strings.HasPrefix(headerValue, prefix) {
		return ""
	}

	return strings.TrimSpace(strings.TrimPrefix(headerValue, prefix))
}

// AccessTokenClaimsAdapter keeps middleware decoupled from the concrete auth package.
type AccessTokenClaimsAdapter struct {
	// ParseFunc validates a token and extracts the trusted user id.
	ParseFunc func(token string, now time.Time) (string, error)
}

// ParseAccessToken validates the token through the adapter callback.
func (a AccessTokenClaimsAdapter) ParseAccessToken(token string, now time.Time) (struct{ UserID string }, error) {
	if a.ParseFunc == nil {
		return struct{ UserID string }{}, errors.New("token parser is not configured")
	}

	userID, err := a.ParseFunc(token, now)
	if err != nil {
		return struct{ UserID string }{}, err
	}

	return struct{ UserID string }{UserID: userID}, nil
}
