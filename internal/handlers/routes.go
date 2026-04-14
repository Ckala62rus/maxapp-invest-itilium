package handlers

import (
	"net/http"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Routes wires middleware and endpoints into the main router.
func (h *Handler) Routes() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.CORS)
	router.Use(middleware.RequestID)
	router.Use(middleware.Identity)
	router.Use(middleware.Recover(h.logger))
	router.Use(middleware.Metrics)
	router.Use(middleware.Logging(h.logger))

	router.Get("/healthz", h.Health)
	router.Get("/readyz", h.Ready)
	router.Handle("/metrics", promhttp.Handler())

	router.Route("/api/v1", func(router chi.Router) {
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

	return router
}
