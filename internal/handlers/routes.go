package handlers

import (
	"net/http"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Routes wires middleware and endpoints into the main router.
func (h *Handler) Routes() http.Handler {
	router := chi.NewRouter()

	// Порядок важен: CORS → request id → кто пользователь → паника → метрики → лог.
	router.Use(middleware.CORS)
	router.Use(middleware.RequestID)
	router.Use(middleware.Identity(h.logger, middleware.AccessTokenClaimsAdapter{
		ParseFunc: func(token string, now time.Time) (string, error) {
			claims, err := h.authManager.ParseAccessToken(token, now)
			if err != nil {
				return "", err
			}

			return claims.UserID, nil
		},
	}, h.debugIdentity))
	router.Use(middleware.Recover(h.logger))
	router.Use(middleware.Metrics)
	router.Use(middleware.Logging(h.logger))

	router.Get("/healthz", h.Health)
	router.Get("/readyz", h.Ready)
	router.Handle("/metrics", promhttp.Handler())

	router.Route("/api/v1", func(router chi.Router) {
		// Обмен MAX initData на backend access token — без RequireIdentity (токена ещё нет).
		router.Post("/auth/max/validate", h.ValidateMaxAuth)

		router.Group(func(router chi.Router) {
			// Все методы ниже требуют непустой userId в контексте (Bearer или debug).
			router.Use(middleware.RequireIdentity)
			router.Get("/users/me", h.GetProfile)
			router.Post("/users/register", h.RegisterUser)
			router.Post("/users/employee", h.FindEmployeeByIdentifier)
			router.Get("/tickets", h.ListMyTickets)
			router.Get("/tickets/responsible", h.ListResponsibleTickets)
			router.Post("/tickets/search", h.SearchTicket)
			router.Post("/tickets", h.CreateTicket)
			router.Get("/tickets/{number}", h.GetTicket)
			router.Post("/tickets/{number}/comments", h.AddComment)
			router.Post("/tickets/{number}/status", h.ChangeStatus)
			router.Get("/tickets/{number}/responsibles", h.ListResponsibleOptions)
			router.Post("/tickets/{number}/responsible", h.ChangeResponsible)
		})
	})

	return router
}
