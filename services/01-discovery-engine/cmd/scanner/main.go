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

	"github.com/ncsound919/modernization-control-plane/services/discovery-engine/internal/api"
	"github.com/ncsound919/modernization-control-plane/services/discovery-engine/internal/scanner"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))

	cfg := scanner.Config{
		Neo4jURI:      getenv("NEO4J_URI", "bolt://localhost:7687"),
		Neo4jUser:     getenv("NEO4J_USER", "neo4j"),
		Neo4jPassword: getenv("NEO4J_PASSWORD", "password"),
		PostgresDSN:   getenv("POSTGRES_DSN", "postgres://mcp:mcp_dev_password@localhost:5432/mcp"),
		Port:          getenv("PORT", "8080"),
	}

	logger.Info("starting discovery-engine",
		"neo4j_uri", cfg.Neo4jURI,
		"port", cfg.Port,
	)

	sc := scanner.New(cfg)
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      api.New(sc, logger),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start HTTP server in background.
	go func() {
		logger.Info("listening", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("server error", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for termination signal then gracefully shut down.
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

func getenv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
