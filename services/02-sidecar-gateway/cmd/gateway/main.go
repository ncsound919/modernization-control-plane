package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/ncsound919/modernization-control-plane/services/sidecar-gateway/internal/api"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	srv := api.New()

	addr := fmt.Sprintf(":%s", port)
	log.Printf("sidecar-gateway starting on %s", addr)
	if err := http.ListenAndServe(addr, srv); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
