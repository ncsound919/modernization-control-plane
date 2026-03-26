package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	coboladapter "github.com/ncsound919/modernization-control-plane/services/sidecar-gateway/internal/adapters/cobol"
	hl7adapter "github.com/ncsound919/modernization-control-plane/services/sidecar-gateway/internal/adapters/hl7"
	"github.com/ncsound919/modernization-control-plane/services/sidecar-gateway/internal/models"
)

// Server is the Sidecar Gateway HTTP API server.
type Server struct {
	mux          *http.ServeMux
	cobol        *coboladapter.Adapter
	hl7          *hl7adapter.Adapter
	requestCount atomic.Int64
	errorCount   atomic.Int64
	totalLatency atomic.Int64 // cumulative ms
}

// New creates and returns a configured Server with all routes registered.
func New() *Server {
	s := &Server{
		mux:   http.NewServeMux(),
		cobol: coboladapter.New(),
		hl7:   hl7adapter.New(),
	}
	s.registerRoutes()
	return s
}

// ServeHTTP makes Server implement http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	s.requestCount.Add(1)

	lw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
	s.mux.ServeHTTP(lw, r)

	latency := time.Since(start).Milliseconds()
	s.totalLatency.Add(latency)
	if lw.status >= 400 {
		s.errorCount.Add(1)
	}

	log.Printf("%s %s %d %dms", r.Method, r.URL.Path, lw.status, latency)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)
	s.mux.HandleFunc("GET /api/v1/adapters", s.handleListAdapters)
	s.mux.HandleFunc("POST /api/v1/adapters/cobol/execute", s.handleCOBOLExecute)
	s.mux.HandleFunc("POST /api/v1/adapters/hl7/transform", s.handleHL7Transform)
	s.mux.HandleFunc("POST /api/v1/adapters/sftp/transfer", s.handleSFTPTransfer)
	s.mux.HandleFunc("GET /api/v1/contracts", s.handleListContracts)
	s.mux.HandleFunc("GET /api/v1/metrics", s.handleMetrics)
	// Catch-all for /api/v1/proxy/{service}/{path...}
	s.mux.HandleFunc("/api/v1/proxy/", s.handleProxy)
}

// -------------------------------------------------------------------------
// Handlers
// -------------------------------------------------------------------------

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	jsonOK(w, map[string]interface{}{
		"status":    "healthy",
		"service":   "sidecar-gateway",
		"version":   "1.0.0",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleListAdapters(w http.ResponseWriter, _ *http.Request) {
	adapters := []models.Adapter{
		{ID: "cobol-mq-01", Name: "COBOL/MQ Batch Adapter", Type: "batch", Status: "active", Protocol: "JCL/MQ"},
		{ID: "ibmmq-01", Name: "IBM MQ Message Adapter", Type: "messaging", Status: "active", Protocol: "IBM MQ"},
		{ID: "sftp-01", Name: "SFTP Flat-File Adapter", Type: "file", Status: "active", Protocol: "SFTP"},
		{ID: "hl7-fhir-01", Name: "HL7 v2 / FHIR Adapter", Type: "healthcare", Status: "active", Protocol: "HL7v2/FHIR"},
		{ID: "db2-vsam-01", Name: "DB2/VSAM Query Adapter", Type: "database", Status: "degraded", Protocol: "DB2/VSAM"},
	}
	jsonOK(w, map[string]interface{}{
		"adapters": adapters,
		"total":    len(adapters),
	})
}

func (s *Server) handleCOBOLExecute(w http.ResponseWriter, r *http.Request) {
	var job models.COBOLJob
	if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := s.cobol.Execute(job)
	if err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	jsonOK(w, result)
}

func (s *Server) handleHL7Transform(w http.ResponseWriter, r *http.Request) {
	var msg models.HL7Message
	if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}

	result, err := s.hl7.Transform(msg)
	if err != nil {
		jsonError(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	jsonOK(w, result)
}

func (s *Server) handleSFTPTransfer(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Host        string `json:"host"`
		Port        int    `json:"port"`
		Username    string `json:"username"`
		RemotePath  string `json:"remote_path"`
		LocalPath   string `json:"local_path"`
		Direction   string `json:"direction"` // "upload" | "download"
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.Host == "" || req.RemotePath == "" {
		jsonError(w, http.StatusBadRequest, "host and remote_path are required")
		return
	}
	if req.Direction == "" {
		req.Direction = "download"
	}
	if req.Port == 0 {
		req.Port = 22
	}

	jsonOK(w, map[string]interface{}{
		"transfer_id":  fmt.Sprintf("xfr-%d", time.Now().UnixMilli()),
		"status":       "queued",
		"host":         req.Host,
		"port":         req.Port,
		"remote_path":  req.RemotePath,
		"direction":    req.Direction,
		"queued_at":    time.Now().UTC().Format(time.RFC3339),
		"message":      "SFTP transfer queued — adapter will process asynchronously",
	})
}

func (s *Server) handleListContracts(w http.ResponseWriter, _ *http.Request) {
	contracts := []models.Contract{
		{ID: "ctr-001", Name: "Accounts API", Version: "v1.2.0", Spec: "openapi-3.1", Backend: "cobol-mq-01"},
		{ID: "ctr-002", Name: "Claims API", Version: "v2.0.0", Spec: "openapi-3.1", Backend: "db2-vsam-01"},
		{ID: "ctr-003", Name: "Patient API", Version: "v1.0.0", Spec: "fhir-r4", Backend: "hl7-fhir-01"},
		{ID: "ctr-004", Name: "Orders API", Version: "v1.1.0", Spec: "grpc/protobuf", Backend: "ibmmq-01"},
		{ID: "ctr-005", Name: "Files API", Version: "v1.0.0", Spec: "openapi-3.1", Backend: "sftp-01"},
	}
	jsonOK(w, map[string]interface{}{
		"contracts": contracts,
		"total":     len(contracts),
	})
}

func (s *Server) handleProxy(w http.ResponseWriter, r *http.Request) {
	// Path: /api/v1/proxy/{service}/{path...}
	trimmed := strings.TrimPrefix(r.URL.Path, "/api/v1/proxy/")
	parts := strings.SplitN(trimmed, "/", 2)
	service := ""
	path := "/"
	if len(parts) > 0 {
		service = parts[0]
	}
	if len(parts) > 1 {
		path = "/" + parts[1]
	}

	if service == "" {
		jsonError(w, http.StatusBadRequest, "service name is required in path")
		return
	}

	// Route table: map service names to mock backend URLs.
	backends := map[string]string{
		"accounts": "http://accounts-service:8080",
		"claims":   "http://claims-service:8080",
		"patients": "http://patients-service:8080",
		"orders":   "http://orders-service:8080",
	}

	backend, ok := backends[service]
	if !ok {
		jsonError(w, http.StatusNotFound, fmt.Sprintf("no backend registered for service %q", service))
		return
	}

	jsonOK(w, map[string]interface{}{
		"proxied":      true,
		"service":      service,
		"backend":      backend,
		"path":         path,
		"method":       r.Method,
		"forwarded_at": time.Now().UTC().Format(time.RFC3339),
		"note":         "stub response — production will forward to real backend",
	})
}

func (s *Server) handleMetrics(w http.ResponseWriter, _ *http.Request) {
	reqCount := s.requestCount.Load()
	errCount := s.errorCount.Load()
	totalLat := s.totalLatency.Load()

	var errorRate float64
	if reqCount > 0 {
		errorRate = float64(errCount) / float64(reqCount)
	}
	var avgLatency float64
	if reqCount > 0 {
		avgLatency = float64(totalLat) / float64(reqCount)
	}

	jsonOK(w, models.GatewayMetrics{
		RequestCount: reqCount,
		ErrorRate:    errorRate,
		AvgLatencyMs: avgLatency,
	})
}

// -------------------------------------------------------------------------
// Helpers
// -------------------------------------------------------------------------

func jsonOK(w http.ResponseWriter, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(v)
}

func jsonError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

// statusWriter wraps ResponseWriter to capture the HTTP status code.
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}
