package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ckala62rus/maxapp-invest-itilium/internal/config"
	"github.com/Ckala62rus/maxapp-invest-itilium/internal/container"
	applogger "github.com/Ckala62rus/maxapp-invest-itilium/internal/logger"
)

func main() {
	// CONFIG_PATH переопределяют в Docker; локально достаточно YAML из deploy/config.
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configPath = "deploy/config/app.dev.yml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := applogger.New(cfg.Logging)
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	app, err := container.Build(ctx, cfg, logger)
	if err != nil {
		logger.Error("build container", "error", err)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:         cfg.HTTP.Address(),
		Handler:      app.Handler.Routes(),
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}

	go func() {
		logger.Info("http server started", "address", cfg.HTTP.Address(), "env", cfg.App.Env, "version", cfg.App.Version)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("http server stopped unexpectedly", "error", err)
			stop()
		}
	}()

	// Ожидаем SIGINT/SIGTERM, затем мягко закрываем listener.
	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("http server shutdown failed", "error", err)
		os.Exit(1)
	}

	logger.Info("http server stopped gracefully")
}
