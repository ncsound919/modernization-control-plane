package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/ncsound919/modernization-control-plane/services/governance-engine/internal/api"
)

func main() {
	port := envOr("PORT", "8080")
	// POSTGRES_DSN and KAFKA_BROKERS are used for future PostgreSQL / Kafka integrations.
	postgresDSN := envOr("POSTGRES_DSN", "postgres://localhost:5432/mcp")
	kafkaBrokers := envOr("KAFKA_BROKERS", "localhost:9092")

	log.Printf("governance-engine starting")
	log.Printf("  port=%s", port)
	log.Printf("  postgres_dsn=%s", redactDSN(postgresDSN))
	log.Printf("  kafka_brokers=%s", kafkaBrokers)

	srv := api.NewServer()

	addr := fmt.Sprintf(":%s", port)
	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server exited: %v", err)
	}
}

func envOr(key, fallback string) string {
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
