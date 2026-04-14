package handlers

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutesCORSPreflight(t *testing.T) {
	t.Parallel()

	handler := New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil)
	router := handler.Routes()

	request := httptest.NewRequest(http.MethodOptions, "/api/v1/tickets", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "Authorization")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow origin header to be %q, got %q", "http://localhost:5173", got)
	}
}

func TestRoutesCORSActualRequest(t *testing.T) {
	t.Parallel()

	handler := New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil)
	router := handler.Routes()

	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "http://localhost:5173")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow origin header to be %q, got %q", "http://localhost:5173", got)
	}
}
