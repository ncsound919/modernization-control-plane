// Package api provides the HTTP server and route handlers for the Discovery Engine.
package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/discovery-engine/internal/models"
	"github.com/ncsound919/modernization-control-plane/services/discovery-engine/internal/scanner"
)

// Server is the HTTP API server for the Discovery Engine.
type Server struct {
	scanner *scanner.Scanner
	mux     *http.ServeMux
	logger  *slog.Logger
}

// New constructs a Server and registers all routes.
func New(sc *scanner.Scanner, logger *slog.Logger) *Server {
	s := &Server{
		scanner: sc,
		mux:     http.NewServeMux(),
		logger:  logger,
	}
	s.registerRoutes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	s.mux.ServeHTTP(w, r)
	s.logger.Info("request", "method", r.Method, "path", r.URL.Path,
		"duration_ms", time.Since(start).Milliseconds())
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/v1/assets", s.handleListAssets)
	s.mux.HandleFunc("POST /api/v1/scans", s.handleCreateScan)
	s.mux.HandleFunc("GET /api/v1/scans/{id}", s.handleGetScan)
	s.mux.HandleFunc("GET /api/v1/graph", s.handleGraph)
	s.mux.HandleFunc("GET /api/v1/certificates", s.handleListCertificates)
}

// handleHealth returns service health and basic connectivity status.
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respond(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "discovery-engine",
		"time":    time.Now().UTC().Format(time.RFC3339),
	})
}

// handleListAssets returns all discovered assets, optionally filtered.
func (s *Server) handleListAssets(w http.ResponseWriter, r *http.Request) {
	assets := s.scanner.Assets()

	// Optional query filter: ?type=legacy
	if t := r.URL.Query().Get("type"); t != "" {
		var filtered []models.Asset
		for _, a := range assets {
			if strings.EqualFold(string(a.Type), t) {
				filtered = append(filtered, a)
			}
		}
		assets = filtered
	}

	// Optional query filter: ?env=aws
	if env := r.URL.Query().Get("env"); env != "" {
		var filtered []models.Asset
		for _, a := range assets {
			if strings.EqualFold(string(a.Environment), env) {
				filtered = append(filtered, a)
			}
		}
		assets = filtered
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"total":  len(assets),
		"assets": assets,
	})
}

// handleCreateScan triggers a new asynchronous scan.
func (s *Server) handleCreateScan(w http.ResponseWriter, r *http.Request) {
	var req models.ScanRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}
	if len(req.ScanTypes) == 0 {
		req.ScanTypes = []string{"certs", "assets", "legacy"}
	}
	if req.Environment == "" {
		req.Environment = models.EnvOnPrem
	}

	scan := s.scanner.StartScan(r.Context(), req)
	respond(w, http.StatusAccepted, scan)
}

// handleGetScan returns the status of a specific scan.
func (s *Server) handleGetScan(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	scan, ok := s.scanner.GetScan(id)
	if !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("scan %q not found", id))
		return
	}
	respond(w, http.StatusOK, scan)
}

// handleGraph returns the full discovery graph.
func (s *Server) handleGraph(w http.ResponseWriter, r *http.Request) {
	scanID := r.URL.Query().Get("scan_id")
	if scanID == "" {
		scanID = "latest"
	}
	graph := s.scanner.BuildGraph(scanID)
	respond(w, http.StatusOK, graph)
}

// handleListCertificates returns all discovered certificates with optional risk filters.
func (s *Server) handleListCertificates(w http.ResponseWriter, r *http.Request) {
	certs := s.scanner.Certificates()

	// Optional filter: ?expiring_soon=true
	if r.URL.Query().Get("expiring_soon") == "true" {
		var filtered []models.Certificate
		for _, c := range certs {
			if c.IsExpiringSoon {
				filtered = append(filtered, c)
			}
		}
		certs = filtered
	}

	// Optional filter: ?expired=true
	if r.URL.Query().Get("expired") == "true" {
		var filtered []models.Certificate
		for _, c := range certs {
			if c.IsExpired {
				filtered = append(filtered, c)
			}
		}
		certs = filtered
	}

	respond(w, http.StatusOK, map[string]interface{}{
		"total":        len(certs),
		"certificates": certs,
	})
}

// respond writes a JSON response with the given status code.
func respond(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("failed to encode response", "error", err)
	}
}

// respondError writes a JSON error response.
func respondError(w http.ResponseWriter, status int, msg string) {
	respond(w, status, map[string]string{"error": msg})
}
