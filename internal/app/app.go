package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"taifa-exchange/internal/config"
	"taifa-exchange/internal/platform/httpserver"
)

type App struct {
	cfg        config.Config
	logger     *slog.Logger
	httpServer *http.Server
}

func New(cfg config.Config, logger *slog.Logger) (*App, error) {
	if logger == nil {
		logger = slog.Default()
	}

	router := chi.NewRouter()

	router.Use(httpserver.CorrelationIDMiddleware)
	router.Use(httpserver.RequestLoggerMiddleware(logger))
	router.Use(httpserver.RecovererMiddleware(logger))

	router.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		httpserver.WriteJSON(w, r, http.StatusOK, map[string]any{
			"correlation_id": httpserver.CorrelationIDFromContext(r.Context()),
			"status":         "ok",
			"service":        cfg.ServiceName,
			"environment":    cfg.Environment,
		})
	})

	router.Get("/readyz", func(w http.ResponseWriter, r *http.Request) {
		httpserver.WriteJSON(w, r, http.StatusOK, map[string]any{
			"correlation_id": httpserver.CorrelationIDFromContext(r.Context()),
			"status":         "ok",
			"service":        cfg.ServiceName,
			"dependencies": map[string]string{
				"database": "not_configured",
				"taifa_id": "not_configured",
			},
		})
	})

	server := httpserver.New(httpserver.Config{
		Addr:         cfg.HTTP.Addr,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
	}, router)

	return &App{
		cfg:        cfg,
		logger:     logger,
		httpServer: server,
	}, nil
}

func (a *App) Run(ctx context.Context) error {
	if a == nil || a.httpServer == nil {
		return fmt.Errorf("application is not initialized")
	}

	serverErrors := make(chan error, 1)

	go func() {
		a.logger.Info(
			"starting HTTP server",
			"service", a.cfg.ServiceName,
			"environment", a.cfg.Environment,
			"addr", a.cfg.HTTP.Addr,
		)

		err := a.httpServer.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErrors <- err
			return
		}

		serverErrors <- nil
	}()

	select {
	case <-ctx.Done():
		a.logger.Info("shutdown signal received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), a.cfg.HTTP.ShutdownTimeout)
		defer cancel()

		if err := a.httpServer.Shutdown(shutdownCtx); err != nil {
			return err
		}

		a.logger.Info("HTTP server stopped cleanly")
		return nil

	case err := <-serverErrors:
		return err
	}
}
