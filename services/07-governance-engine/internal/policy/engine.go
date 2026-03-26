package policy

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/governance-engine/internal/models"
)

// Engine evaluates governance policies against a request context.
// In production this would delegate to an embedded OPA instance; here
// the Rego policies are expressed as equivalent Go rule-sets so the
// service compiles and runs with the standard library only.
type Engine struct {
	mu       sync.RWMutex
	policies map[string]*models.Policy
}

func NewEngine() *Engine {
	e := &Engine{
		policies: make(map[string]*models.Policy),
	}
	e.seedBuiltins()
	return e
}

// seedBuiltins loads the built-in compliance policy stubs.
func (e *Engine) seedBuiltins() {
	now := time.Now().UTC()
	builtins := []*models.Policy{
		{
			ID:          "hipaa-phi-access",
			Name:        "HIPAA PHI Access Control",
			Description: "PHI data requires explicit human approval before access",
			Type:        "deny",
			Framework:   models.FrameworkHIPAA,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "hipaa-audit-required",
			Name:        "HIPAA Audit Logging Required",
			Description: "All operations on PHI must be audit-logged",
			Type:        "deny",
			Framework:   models.FrameworkHIPAA,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "gdpr-consent-required",
			Name:        "GDPR Consent Required",
			Description: "Personal data processing requires valid consent",
			Type:        "deny",
			Framework:   models.FrameworkGDPR,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "gdpr-data-minimisation",
			Name:        "GDPR Data Minimisation",
			Description: "Only data necessary for the stated purpose may be processed",
			Type:        "deny",
			Framework:   models.FrameworkGDPR,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "soc2-least-privilege",
			Name:        "SOC 2 Least Privilege",
			Description: "Access must be limited to the minimum required for the task",
			Type:        "deny",
			Framework:   models.FrameworkSOC2,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "soc2-mfa-required",
			Name:        "SOC 2 MFA Required",
			Description: "Multi-factor authentication required for privileged operations",
			Type:        "deny",
			Framework:   models.FrameworkSOC2,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "soc2-cost-cap",
			Name:        "SOC 2 Workflow Cost Cap",
			Description: "Workflow cost must not exceed the tenant cost limit",
			Type:        "deny",
			Framework:   models.FrameworkSOC2,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:          "pcidss-cardholder-data",
			Name:        "PCI-DSS Cardholder Data Protection",
			Description: "Cardholder data must not be stored in plaintext",
			Type:        "deny",
			Framework:   models.FrameworkPCIDSS,
			Enabled:     true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}
	for _, p := range builtins {
		e.policies[p.ID] = p
	}
}

// ListPolicies returns all registered policies.
func (e *Engine) ListPolicies() []*models.Policy {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]*models.Policy, 0, len(e.policies))
	for _, p := range e.policies {
		cp := *p
		out = append(out, &cp)
	}
	return out
}

// AddPolicy registers a new policy.
func (e *Engine) AddPolicy(p *models.Policy) {
	e.mu.Lock()
	defer e.mu.Unlock()
	now := time.Now().UTC()
	p.CreatedAt = now
	p.UpdatedAt = now
	e.policies[p.ID] = p
}

// Evaluate runs all enabled policies for the given framework (or all if empty)
// against the provided input context.
func (e *Engine) Evaluate(req *models.PolicyEvaluationRequest) *models.PolicyDecision {
	e.mu.RLock()
	defer e.mu.RUnlock()
	decision := &models.PolicyDecision{
		Allowed:     true,
		EvaluatedAt: time.Now().UTC(),
		Framework:   req.Framework,
	}

	var violations []string

	for _, p := range e.policies {
		if !p.Enabled {
			continue
		}
		if req.Framework != "" && p.Framework != req.Framework {
			continue
		}
		if req.PolicyID != "" && p.ID != req.PolicyID {
			continue
		}

		msgs := e.evalPolicy(p, req.Input)
		violations = append(violations, msgs...)
	}

	if len(violations) > 0 {
		decision.Allowed = false
		decision.Reason = strings.Join(violations, "; ")
		decision.Violations = violations
	} else {
		decision.Reason = "all applicable policies passed"
	}

	return decision
}

// evalPolicy applies a single policy's rules to the input context and returns
// any violation messages (empty slice means allowed).
func (e *Engine) evalPolicy(p *models.Policy, input map[string]interface{}) []string {
	var msgs []string

	str := func(key string) string {
		if v, ok := input[key]; ok {
			return fmt.Sprintf("%v", v)
		}
		return ""
	}
	boolVal := func(key string) bool {
		if v, ok := input[key]; ok {
			if b, ok := v.(bool); ok {
				return b
			}
		}
		return false
	}
	floatVal := func(key string) float64 {
		if v, ok := input[key]; ok {
			switch n := v.(type) {
			case float64:
				return n
			case int:
				return float64(n)
			}
		}
		return 0
	}

	switch p.ID {
	case "hipaa-phi-access":
		if strings.EqualFold(str("data_classification"), "phi") &&
			str("action") == "read" && !boolVal("human_approved") {
			msgs = append(msgs, "HIPAA: PHI access requires explicit human approval")
		}

	case "hipaa-audit-required":
		if strings.EqualFold(str("data_classification"), "phi") && !boolVal("audit_logged") {
			msgs = append(msgs, "HIPAA: all PHI operations must be audit-logged")
		}

	case "gdpr-consent-required":
		classification := strings.ToLower(str("data_classification"))
		if (classification == "personal" || classification == "pii") &&
			!boolVal("consent_given") && str("lawful_basis") == "" {
			msgs = append(msgs, "GDPR: processing personal data requires valid consent or another lawful basis")
		}

	case "gdpr-data-minimisation":
		if boolVal("excess_data_requested") {
			msgs = append(msgs, "GDPR: only data necessary for the stated purpose may be processed")
		}

	case "soc2-least-privilege":
		role := strings.ToLower(str("actor_role"))
		action := strings.ToLower(str("action"))
		if role == "guest" && (action == "write" || action == "delete" || action == "admin") {
			msgs = append(msgs, "SOC 2: guest role may not perform privileged operations")
		}

	case "soc2-mfa-required":
		action := strings.ToLower(str("action"))
		privileged := action == "delete" || action == "admin" || action == "write_sensitive"
		if privileged && !boolVal("mfa_verified") {
			msgs = append(msgs, "SOC 2: multi-factor authentication required for privileged operations")
		}

	case "soc2-cost-cap":
		workflowCost := floatVal("workflow_cost_usd")
		costLimit := floatVal("tenant_cost_limit_usd")
		if costLimit > 0 && workflowCost > costLimit {
			msgs = append(msgs, fmt.Sprintf("SOC 2: workflow cost cap exceeded ($%.2f > $%.2f)", workflowCost, costLimit))
		}

	case "pcidss-cardholder-data":
		if strings.EqualFold(str("data_classification"), "cardholder") &&
			strings.EqualFold(str("storage_format"), "plaintext") {
			msgs = append(msgs, "PCI-DSS: cardholder data must not be stored in plaintext")
		}

	default:
		// Custom / unknown policy — allow by default.
	}

	return msgs
}

// ComplianceStatus returns a compliance summary for each framework.
func (e *Engine) ComplianceStatus() []*models.ComplianceStatus {
	e.mu.RLock()
	defer e.mu.RUnlock()
	now := time.Now().UTC()
	frameworks := []models.PolicyFramework{
		models.FrameworkHIPAA,
		models.FrameworkGDPR,
		models.FrameworkSOC2,
		models.FrameworkPCIDSS,
	}

	var statuses []*models.ComplianceStatus
	for _, fw := range frameworks {
		enabled := 0
		for _, p := range e.policies {
			if p.Framework == fw && p.Enabled {
				enabled++
			}
		}
		status := "compliant"
		score := 100
		if enabled == 0 {
			status = "unknown"
			score = 0
		}
		statuses = append(statuses, &models.ComplianceStatus{
			Framework:   fw,
			Status:      status,
			Score:       score,
			Violations:  []models.ComplianceViolation{},
			LastChecked: now,
		})
	}
	return statuses
}
