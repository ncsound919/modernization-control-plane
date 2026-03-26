package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/governance-engine/internal/audit"
	"github.com/ncsound919/modernization-control-plane/services/governance-engine/internal/killswitch"
	"github.com/ncsound919/modernization-control-plane/services/governance-engine/internal/models"
	"github.com/ncsound919/modernization-control-plane/services/governance-engine/internal/policy"
)

// Server bundles all governance subsystems behind an HTTP API.
type Server struct {
	mux        *http.ServeMux
	policyEng  *policy.Engine
	auditLog   *audit.Log
	killMgr    *killswitch.Manager
	startedAt  time.Time
}

func NewServer() *Server {
	s := &Server{
		mux:       http.NewServeMux(),
		policyEng: policy.NewEngine(),
		auditLog:  audit.NewLog(),
		killMgr:   killswitch.NewManager(),
		startedAt: time.Now().UTC(),
	}
	s.routes()
	return s
}

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.mux.ServeHTTP(w, r)
}

// routes wires all HTTP endpoints.
func (s *Server) routes() {
	s.mux.HandleFunc("GET /health", s.handleHealth)

	s.mux.HandleFunc("POST /api/v1/policy/evaluate", s.handlePolicyEvaluate)

	s.mux.HandleFunc("GET /api/v1/policies", s.handleListPolicies)
	s.mux.HandleFunc("POST /api/v1/policies", s.handleCreatePolicy)

	s.mux.HandleFunc("GET /api/v1/killswitches", s.handleListKillSwitches)
	// Pattern-based path parameters require manual prefix matching on Go < 1.22.
	// Go 1.22 ServeMux supports {name} wildcards.
	s.mux.HandleFunc("POST /api/v1/killswitches/{name}/activate", s.handleActivateKillSwitch)
	s.mux.HandleFunc("POST /api/v1/killswitches/{name}/deactivate", s.handleDeactivateKillSwitch)

	s.mux.HandleFunc("GET /api/v1/audit", s.handleGetAudit)
	s.mux.HandleFunc("POST /api/v1/audit", s.handleWriteAudit)

	s.mux.HandleFunc("GET /api/v1/compliance/status", s.handleComplianceStatus)
}

// ---------- helpers ----------

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

func errResponse(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

// newID returns a simple timestamp-based unique identifier.
func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}

// ---------- handlers ----------

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":     "healthy",
		"service":    "governance-engine",
		"started_at": s.startedAt,
		"uptime_sec": int(time.Since(s.startedAt).Seconds()),
	})
}

func (s *Server) handlePolicyEvaluate(w http.ResponseWriter, r *http.Request) {
	var req models.PolicyEvaluationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResponse(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if req.Input == nil {
		req.Input = map[string]interface{}{}
	}

	decision := s.policyEng.Evaluate(&req)

	// Write an audit entry for the evaluation.
	_ = s.auditLog.Append(&models.AuditEntry{
		ID:       newID("audit"),
		Actor:    r.Header.Get("X-Actor"),
		Action:   "policy.evaluate",
		Resource: fmt.Sprintf("framework:%s", req.Framework),
		TenantID: req.TenantID,
		Decision: map[bool]string{true: "allow", false: "deny"}[decision.Allowed],
	})

	writeJSON(w, http.StatusOK, decision)
}

func (s *Server) handleListPolicies(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"policies": s.policyEng.ListPolicies(),
	})
}

func (s *Server) handleCreatePolicy(w http.ResponseWriter, r *http.Request) {
	var p models.Policy
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		errResponse(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if p.ID == "" {
		p.ID = newID("policy")
	}
	if p.Name == "" {
		errResponse(w, http.StatusBadRequest, "policy name is required")
		return
	}

	s.policyEng.AddPolicy(&p)

	_ = s.auditLog.Append(&models.AuditEntry{
		ID:       newID("audit"),
		Actor:    r.Header.Get("X-Actor"),
		Action:   "policy.create",
		Resource: "policy:" + p.ID,
		TenantID: p.TenantID,
		Decision: "allow",
	})

	writeJSON(w, http.StatusCreated, &p)
}

func (s *Server) handleListKillSwitches(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"kill_switches": s.killMgr.List(),
	})
}

func (s *Server) handleActivateKillSwitch(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		errResponse(w, http.StatusBadRequest, "kill switch name is required")
		return
	}

	var body struct {
		Actor  string `json:"actor"`
		Reason string `json:"reason"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Actor == "" {
		body.Actor = r.Header.Get("X-Actor")
	}

	ks, err := s.killMgr.Activate(name, body.Actor, body.Reason)
	if err != nil {
		errResponse(w, http.StatusNotFound, err.Error())
		return
	}

	_ = s.auditLog.Append(&models.AuditEntry{
		ID:       newID("audit"),
		Actor:    body.Actor,
		Action:   "killswitch.activate",
		Resource: "killswitch:" + name,
		Decision: "allow",
		Details:  map[string]interface{}{"reason": body.Reason},
	})

	writeJSON(w, http.StatusOK, ks)
}

func (s *Server) handleDeactivateKillSwitch(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if name == "" {
		errResponse(w, http.StatusBadRequest, "kill switch name is required")
		return
	}

	var body struct {
		Actor string `json:"actor"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if body.Actor == "" {
		body.Actor = r.Header.Get("X-Actor")
	}

	ks, err := s.killMgr.Deactivate(name, body.Actor)
	if err != nil {
		errResponse(w, http.StatusNotFound, err.Error())
		return
	}

	_ = s.auditLog.Append(&models.AuditEntry{
		ID:       newID("audit"),
		Actor:    body.Actor,
		Action:   "killswitch.deactivate",
		Resource: "killswitch:" + name,
		Decision: "allow",
	})

	writeJSON(w, http.StatusOK, ks)
}

func (s *Server) handleGetAudit(w http.ResponseWriter, r *http.Request) {
	entries := s.auditLog.Entries()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}

func (s *Server) handleWriteAudit(w http.ResponseWriter, r *http.Request) {
	var entry models.AuditEntry
	if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
		errResponse(w, http.StatusBadRequest, "invalid request body: "+err.Error())
		return
	}
	if entry.ID == "" {
		entry.ID = newID("audit")
	}
	if entry.Actor == "" {
		entry.Actor = r.Header.Get("X-Actor")
	}

	if err := s.auditLog.Append(&entry); err != nil {
		errResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, &entry)
}

func (s *Server) handleComplianceStatus(w http.ResponseWriter, r *http.Request) {
	statuses := s.policyEng.ComplianceStatus()
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"frameworks": statuses,
		"checked_at": time.Now().UTC(),
	})
}
