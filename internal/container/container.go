// Package container builds the application dependency graph.
package container

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/api"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/config"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/handlers"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/repository"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/services"
	"github.com/redis/go-redis/v9"
)

// Container keeps long-lived dependencies in one composition root.
type Container struct {
	// Config stores the loaded runtime configuration.
	Config *config.Config
	// Logger stores the shared structured logger.
	Logger *slog.Logger
	// Handler exposes the fully wired HTTP handler bundle.
	Handler *handlers.Handler
	// Redis stores the optional low-level Redis client.
	Redis *redis.Client
}

// Build creates repositories, services and handlers in one place.
func Build(ctx context.Context, cfg *config.Config, logger *slog.Logger) (*Container, error) {
	var redisClient *redis.Client
	if cfg.Redis.Enabled {
		redisClient = redis.NewClient(&redis.Options{
			Addr:     cfg.Redis.Address,
			Password: cfg.Redis.Password,
			DB:       cfg.Redis.DB,
		})

		pingCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()

		if err := redisClient.Ping(pingCtx).Err(); err != nil {
			logger.Warn("redis ping failed, continuing without cache", "error", err)
			redisClient = nil
		} else {
			logger.Info("redis connected", "address", cfg.Redis.Address, "db", cfg.Redis.DB)
		}
	}

	cache := repository.NewRedisCache(redisClient)
	profileRepository := repository.NewMemoryUserRepository()

	var itiliumClient services.ItiliumClient
	if cfg.App.DemoMode {
		logger.Info("using demo itilium client")
		itiliumClient = api.NewDemoClient()
	} else {
		if cfg.Itilium.BaseURL == "" {
			return nil, fmt.Errorf("itilium base url is required when demo mode is disabled")
		}
		itiliumClient = api.NewClient(cfg.Itilium, logger)
	}

	profileService := services.NewProfileService(profileRepository)
	ticketService := services.NewTicketService(itiliumClient, cache)
	handler := handlers.New(logger, profileService, ticketService)

	return &Container{
		Config:  cfg,
		Logger:  logger,
		Handler: handler,
		Redis:   redisClient,
	}, nil
}
