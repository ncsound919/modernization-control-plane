// Package main is the entrypoint for the Certificate Lifecycle Manager (CLM) service.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/api"
	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/inventory"
	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/policy"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	postgresDSN := getenv("POSTGRES_DSN", "postgres://mcp:mcp_dev_password@localhost:5432/mcp")
	vaultAddr := getenv("VAULT_ADDR", "http://localhost:8200")
	port := getenv("PORT", "8080")

	logger.Info("starting clm-service",
		"postgres_dsn", redactDSN(postgresDSN),
		"vault_addr", vaultAddr,
		"port", port,
		"note", "using in-memory store (PostgreSQL/Vault not required on startup)",
	)

	store := inventory.New()
	engine := policy.New(store, logger)

	// Run the policy engine periodically in the background.
	go runScheduler(engine, logger)

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      api.New(store, engine, logger),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("shutdown error", "error", err)
	}
	logger.Info("stopped")
}

// runScheduler runs the policy engine on a fixed interval (60 s). It logs a
// heartbeat on every tick so operators can confirm the scheduler is alive even
// when no certs need rotation.
func runScheduler(engine *policy.Engine, logger *slog.Logger) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// Run once immediately so the first check happens at startup.
	checkAndLog(engine, logger)

	for range ticker.C {
		checkAndLog(engine, logger)
	}
}

func checkAndLog(engine *policy.Engine, logger *slog.Logger) {
	logger.Info("checking certificate rotation policies")
	results := engine.Evaluate()
	if len(results) > 0 {
		logger.Warn("certificates require rotation", "count", len(results))
	}
}

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// redactDSN returns the DSN with the password omitted, leaving only the username.
func redactDSN(dsn string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return "[redacted]"
	}
	if u.User != nil {
		u.User = url.User(u.User.Username())
	}
	return u.String()
}
