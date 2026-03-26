// Package api provides the HTTP server and route handlers for the CLM service.
package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/inventory"
	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/models"
	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/policy"
)

// Server is the HTTP API server for the CLM service.
type Server struct {
	store  *inventory.Store
	engine *policy.Engine
	mux    *http.ServeMux
	logger *slog.Logger
}

// New constructs a Server and registers all routes.
func New(store *inventory.Store, engine *policy.Engine, logger *slog.Logger) *Server {
	s := &Server{
		store:  store,
		engine: engine,
		mux:    http.NewServeMux(),
		logger: logger,
	}
	s.registerRoutes()
	return s
}

// ServeHTTP implements http.Handler.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	s.mux.ServeHTTP(w, r)
	s.logger.Info("request",
		"method", r.Method,
		"path", r.URL.Path,
		"duration_ms", time.Since(start).Milliseconds(),
	)
}

func (s *Server) registerRoutes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)

	// Certificate inventory
	s.mux.HandleFunc("GET /api/v1/certificates", s.handleListCertificates)
	s.mux.HandleFunc("POST /api/v1/certificates", s.handleAddCertificate)
	s.mux.HandleFunc("GET /api/v1/certificates/{id}", s.handleGetCertificate)
	s.mux.HandleFunc("POST /api/v1/certificates/{id}/rotate", s.handleRotateCertificate)
	s.mux.HandleFunc("GET /api/v1/certificates/{id}/status", s.handleGetCertificateStatus)

	// Rotation policies
	s.mux.HandleFunc("GET /api/v1/policies", s.handleListPolicies)
	s.mux.HandleFunc("POST /api/v1/policies", s.handleCreatePolicy)

	// ACME accounts
	s.mux.HandleFunc("GET /api/v1/acme/accounts", s.handleListACMEAccounts)

	// Policy engine evaluation (on-demand)
	s.mux.HandleFunc("POST /api/v1/evaluate", s.handleEvaluate)
}

// handleHealth returns service liveness and basic configuration status.
func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	respond(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "clm-service",
		"time":    time.Now().UTC().Format(time.RFC3339),
		"version": "1.0.0",
	})
}

// handleListCertificates returns the full certificate inventory with optional filters.
func (s *Server) handleListCertificates(w http.ResponseWriter, r *http.Request) {
	certs := s.store.ListCertificates()

	// ?status=expiring_soon|expired|active|renewing|revoked
	if st := r.URL.Query().Get("status"); st != "" {
		var filtered []*models.Certificate
		for _, c := range certs {
			if string(c.Status) == st {
				filtered = append(filtered, c)
			}
		}
		certs = filtered
	}

	// ?auto_renew=true|false
	if ar := r.URL.Query().Get("auto_renew"); ar != "" {
		want := ar == "true"
		var filtered []*models.Certificate
		for _, c := range certs {
			if c.AutoRenew == want {
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

// handleAddCertificate adds a new certificate to the managed inventory.
func (s *Server) handleAddCertificate(w http.ResponseWriter, r *http.Request) {
	var req models.AddCertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}
	if req.Domain == "" {
		respondError(w, http.StatusBadRequest, "domain is required")
		return
	}
	if req.Provider == "" {
		req.Provider = models.ProviderLetsEncrypt
	}
	if req.ChallengeType == "" {
		req.ChallengeType = models.ChallengeDNS01
	}

	now := time.Now().UTC()
	cert := &models.Certificate{
		ID:              s.store.NextCertID(),
		Domain:          req.Domain,
		SANs:            req.SANs,
		Provider:        req.Provider,
		Issuer:          issuerForProvider(req.Provider),
		IssuedAt:        now,
		ExpiresAt:       now.Add(90 * 24 * time.Hour),
		ValidityDays:    90,
		DaysUntilExpiry: 90,
		Status:          models.CertStatusActive,
		AutoRenew:       req.AutoRenew,
		RotationPolicy:  req.PolicyID,
		DeployTargets:   req.DeployTargets,
		Fingerprint:     "pending",
		SerialNumber:    "pending",
		DiscoveredBy:    "clm-api",
	}
	s.store.AddCertificate(cert)
	s.logger.Info("certificate added", "cert_id", cert.ID, "domain", cert.Domain)
	respond(w, http.StatusCreated, cert)
}

// handleGetCertificate returns details for a specific certificate.
func (s *Server) handleGetCertificate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cert, ok := s.store.GetCertificate(id)
	if !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("certificate %q not found", id))
		return
	}
	respond(w, http.StatusOK, cert)
}

// handleRotateCertificate manually triggers certificate rotation.
func (s *Server) handleRotateCertificate(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, ok := s.store.GetCertificate(id)
	if !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("certificate %q not found", id))
		return
	}

	event, err := s.engine.TriggerRotation(id, "manual-api")
	if err != nil {
		respondError(w, http.StatusInternalServerError, fmt.Sprintf("rotation failed: %v", err))
		return
	}
	respond(w, http.StatusAccepted, map[string]interface{}{
		"message": "rotation triggered",
		"event":   event,
	})
}

// handleGetCertificateStatus returns the current status and recent events for a certificate.
func (s *Server) handleGetCertificateStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	cert, ok := s.store.GetCertificate(id)
	if !ok {
		respondError(w, http.StatusNotFound, fmt.Sprintf("certificate %q not found", id))
		return
	}
	events := s.store.ListRotationEvents(id)
	respond(w, http.StatusOK, map[string]interface{}{
		"cert_id":          cert.ID,
		"domain":           cert.Domain,
		"status":           cert.Status,
		"days_until_expiry": cert.DaysUntilExpiry,
		"expires_at":       cert.ExpiresAt,
		"auto_renew":       cert.AutoRenew,
		"rotation_events":  events,
	})
}

// handleListPolicies returns all rotation policies.
func (s *Server) handleListPolicies(w http.ResponseWriter, _ *http.Request) {
	policies := s.store.ListPolicies()
	respond(w, http.StatusOK, map[string]interface{}{
		"total":    len(policies),
		"policies": policies,
	})
}

// handleCreatePolicy creates a new rotation policy.
func (s *Server) handleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, fmt.Sprintf("invalid request body: %v", err))
		return
	}
	if req.Name == "" {
		respondError(w, http.StatusBadRequest, "name is required")
		return
	}
	if req.DaysBeforeExpiry <= 0 {
		req.DaysBeforeExpiry = 30
	}
	if req.Provider == "" {
		req.Provider = models.ProviderLetsEncrypt
	}
	if req.ChallengeType == "" {
		req.ChallengeType = models.ChallengeDNS01
	}

	now := time.Now().UTC()
	pol := &models.RotationPolicy{
		ID:               s.store.NextPolicyID(),
		Name:             req.Name,
		Description:      req.Description,
		DaysBeforeExpiry: req.DaysBeforeExpiry,
		Provider:         req.Provider,
		ChallengeType:    req.ChallengeType,
		DeployTargets:    req.DeployTargets,
		AutoApprove:      req.AutoApprove,
		NotifyEmail:      req.NotifyEmail,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	s.store.AddPolicy(pol)
	s.logger.Info("policy created", "policy_id", pol.ID, "name", pol.Name)
	respond(w, http.StatusCreated, pol)
}

// handleListACMEAccounts returns all registered ACME accounts.
func (s *Server) handleListACMEAccounts(w http.ResponseWriter, _ *http.Request) {
	accounts := s.store.ListACMEAccounts()
	respond(w, http.StatusOK, map[string]interface{}{
		"total":    len(accounts),
		"accounts": accounts,
	})
}

// handleEvaluate runs the policy engine immediately and returns the results.
func (s *Server) handleEvaluate(w http.ResponseWriter, _ *http.Request) {
	results := s.engine.Evaluate()
	respond(w, http.StatusOK, map[string]interface{}{
		"evaluated_at":    time.Now().UTC().Format(time.RFC3339),
		"action_required": len(results),
		"results":         results,
	})
}

// issuerForProvider returns a human-readable issuer name for a given provider.
func issuerForProvider(p models.Provider) string {
	switch p {
	case models.ProviderLetsEncrypt:
		return "Let's Encrypt Authority X3"
	case models.ProviderDigiCert:
		return "DigiCert Global CA G2"
	case models.ProviderSectigo:
		return "Sectigo RSA Domain Validation Secure Server CA"
	case models.ProviderInternal:
		return "Internal EJBCA Root CA"
	default:
		return "Unknown CA"
	}
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
