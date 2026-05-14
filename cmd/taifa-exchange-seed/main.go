package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"taifa-exchange/internal/config"
	"taifa-exchange/internal/platform/postgres"
	"taifa-exchange/internal/policy"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := config.Load()

	if cfg.Database.DSN == "" {
		logger.Error("TAIFA_EXCHANGE_DATABASE_DSN is required")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Seed.Timeout)
	defer cancel()

	pool, err := postgres.Open(ctx, postgres.Config{
		DSN:            cfg.Database.DSN,
		MinConns:       cfg.Database.MinConns,
		MaxConns:       cfg.Database.MaxConns,
		ConnectTimeout: cfg.Database.ConnectTimeout,
	})
	if err != nil {
		logger.Error("failed to open postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	seeder := policy.NewSeeder(pool)

	summary, err := seeder.SeedCanonical(ctx)
	if err != nil {
		logger.Error("failed to seed canonical exchange policies", "error", err)
		os.Exit(1)
	}

	fmt.Printf(
		"taifa-exchange seed completed policies=%d roles=%d capabilities=%d audit_events=%d seed_timeout=%s\n",
		summary.Policies,
		summary.Roles,
		summary.Capabilities,
		summary.AuditEvents,
		cfg.Seed.Timeout,
	)
}
