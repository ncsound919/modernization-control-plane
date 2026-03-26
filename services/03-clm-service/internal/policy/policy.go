// Package policy implements the certificate rotation policy engine.
//
// The engine evaluates every certificate in the inventory against its assigned
// RotationPolicy and flags certificates that should be rotated. It is designed
// to run as a periodic background job so that no certificate expires without a
// renewal attempt being triggered.
package policy

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/inventory"
	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/models"
)

// Engine evaluates rotation policies against the certificate inventory.
type Engine struct {
	store  *inventory.Store
	logger *slog.Logger
}

// New constructs a policy Engine.
func New(store *inventory.Store, logger *slog.Logger) *Engine {
	return &Engine{store: store, logger: logger}
}

// EvaluationResult holds the outcome of a policy check for one certificate.
type EvaluationResult struct {
	CertID          string    `json:"cert_id"`
	Domain          string    `json:"domain"`
	DaysUntilExpiry int       `json:"days_until_expiry"`
	PolicyID        string    `json:"policy_id,omitempty"`
	ShouldRotate    bool      `json:"should_rotate"`
	Reason          string    `json:"reason"`
	EvaluatedAt     time.Time `json:"evaluated_at"`
}

// Evaluate checks every certificate against its policy and returns results for
// any certificate that requires action (rotation or alert).
func (e *Engine) Evaluate() []EvaluationResult {
	certs := e.store.ListCertificates()
	now := time.Now().UTC()
	var results []EvaluationResult

	for _, cert := range certs {
		result := e.evaluateCert(cert, now)
		if result.ShouldRotate {
			results = append(results, result)
			e.logger.Warn("certificate requires rotation",
				"cert_id", cert.ID,
				"domain", cert.Domain,
				"days_until_expiry", result.DaysUntilExpiry,
				"reason", result.Reason,
			)
		}
	}

	e.logger.Info("policy evaluation complete",
		"total_certs", len(certs),
		"action_required", len(results),
	)
	return results
}

// evaluateCert decides whether a single certificate needs rotation.
func (e *Engine) evaluateCert(cert *models.Certificate, now time.Time) EvaluationResult {
	daysUntilExpiry := int(cert.ExpiresAt.Sub(now).Hours() / 24)
	result := EvaluationResult{
		CertID:          cert.ID,
		Domain:          cert.Domain,
		PolicyID:        cert.RotationPolicy,
		DaysUntilExpiry: daysUntilExpiry,
		EvaluatedAt:     now,
	}

	// Already expired — always flag.
	if daysUntilExpiry < 0 {
		result.ShouldRotate = true
		result.Reason = "certificate is expired"
		return result
	}

	// No auto-renew and no policy — nothing to do.
	if !cert.AutoRenew && cert.RotationPolicy == "" {
		result.Reason = "auto-renew disabled, no policy assigned"
		return result
	}

	// Evaluate against the assigned policy if present.
	if cert.RotationPolicy != "" {
		policy, ok := e.store.GetPolicy(cert.RotationPolicy)
		if ok {
			result.PolicyID = policy.ID
			if daysUntilExpiry <= policy.DaysBeforeExpiry {
				result.ShouldRotate = true
				result.Reason = "within rotation window defined by policy"
			}
			return result
		}
	}

	// Fall back to a conservative default (30 days) when no policy is assigned.
	const defaultDaysBeforeExpiry = 30
	if daysUntilExpiry <= defaultDaysBeforeExpiry {
		result.ShouldRotate = true
		result.Reason = "within default 30-day rotation window"
	}
	return result
}

// TriggerRotation marks a certificate as renewing and records a rotation event.
// In a production system this would invoke the ACME client and deploy the new cert.
func (e *Engine) TriggerRotation(certID, triggeredBy string) (*models.CertRotationEvent, error) {
	cert, ok := e.store.GetCertificate(certID)
	if !ok {
		return nil, fmt.Errorf("certificate %s not found", certID)
	}

	eventID := e.store.NextEventID()
	event := &models.CertRotationEvent{
		ID:          eventID,
		CertID:      certID,
		Domain:      cert.Domain,
		Status:      models.RotationEventPending,
		TriggeredBy: triggeredBy,
		StartedAt:   time.Now().UTC(),
	}
	e.store.AddRotationEvent(event)
	e.store.UpdateCertStatus(certID, models.CertStatusRenewing)

	e.logger.Info("rotation triggered",
		"cert_id", certID,
		"domain", cert.Domain,
		"event_id", eventID,
		"triggered_by", triggeredBy,
	)
	return event, nil
}
