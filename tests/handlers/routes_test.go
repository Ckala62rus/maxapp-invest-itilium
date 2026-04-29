// Package handlers_test проверяет HTTP-слой как внешний клиент.
// Здесь мы не поднимаем настоящий сервер на порту, а используем httptest.
package handlers_test

import (
	// io.Discard нужен, чтобы тестовый logger никуда не писал.
	"io"
	"log/slog"
	"net/http"
	// httptest позволяет создать fake request/response без реального TCP-порта.
	"net/http/httptest"
	"testing"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/handlers"
)

func TestRoutesCORSPreflight(t *testing.T) {
	t.Parallel()

	// Создаём handler с nil-зависимостями: этот тест проверяет только routing/CORS,
	// а не бизнес-логику сервисов.
	handler := handlers.New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil, false, nil, nil)
	// Routes() возвращает http.Handler/router, который можно вызвать напрямую.
	router := handler.Routes()

	// OPTIONS + Access-Control-Request-* — это CORS preflight запрос браузера.
	request := httptest.NewRequest(http.MethodOptions, "/api/v1/tickets", nil)
	request.Header.Set("Origin", "http://localhost:5173")
	request.Header.Set("Access-Control-Request-Method", http.MethodGet)
	request.Header.Set("Access-Control-Request-Headers", "Authorization")

	// ResponseRecorder собирает то, что handler записал бы в HTTP-ответ.
	response := httptest.NewRecorder()
	// ServeHTTP — центральный метод любого http.Handler в Go.
	router.ServeHTTP(response, request)

	// Для preflight ожидаем 204 No Content.
	if response.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, response.Code)
	}

	// Проверяем, что CORS middleware разрешил frontend origin.
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow origin header to be %q, got %q", "http://localhost:5173", got)
	}
}

func TestRoutesCORSActualRequest(t *testing.T) {
	t.Parallel()

	handler := handlers.New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil, false, nil, nil)
	router := handler.Routes()

	// /healthz не требует авторизации и удобен для проверки обычного CORS-ответа.
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	request.Header.Set("Origin", "http://localhost:5173")

	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	// Health endpoint должен отвечать 200 OK.
	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}

	// Даже для обычного GET должен быть Access-Control-Allow-Origin.
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:5173" {
		t.Fatalf("expected allow origin header to be %q, got %q", "http://localhost:5173", got)
	}
}

func TestRoutesMarketingServicesUnauthenticatedExpectsUnauthorizedNotNotFound(t *testing.T) {
	t.Parallel()

	handler := handlers.New(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, nil, false, nil, nil)
	router := handler.Routes()

	// Идём в защищённый маркетинговый route без Authorization и без X-User-ID.
	request := httptest.NewRequest(http.MethodGet, "/api/v1/marketing/services", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	// Если получили 404, значит маршрут вообще не зарегистрирован — это регресс routing.
	if response.Code == http.StatusNotFound {
		t.Fatal("marketing route missing: GET /api/v1/marketing/services returned 404; rebuild/restart backend (see docker-compose.dev.yml backend-dev)")
	}
	// Правильный ответ без identity — 401 Unauthorized.
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d without identity, got %d", http.StatusUnauthorized, response.Code)
	}
}
