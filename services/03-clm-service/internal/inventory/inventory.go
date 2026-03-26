// Package inventory manages the in-memory certificate inventory with a stub for PostgreSQL.
package inventory

import (
	"fmt"
	"sync"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/models"
)

// Store holds certificates, policies, ACME accounts, and rotation events.
type Store struct {
	mu       sync.RWMutex
	certs    map[string]*models.Certificate
	policies map[string]*models.RotationPolicy
	accounts map[string]*models.ACMEAccount
	events   map[string]*models.CertRotationEvent
}

// New creates a Store pre-loaded with realistic mock data.
//
// In production this layer connects to PostgreSQL via POSTGRES_DSN. The mock
// data demonstrates certificates at every stage of the validity lifecycle,
// matching the CA/B Forum reduction schedule described in the README.
func New() *Store {
	s := &Store{
		certs:    make(map[string]*models.Certificate),
		policies: make(map[string]*models.RotationPolicy),
		accounts: make(map[string]*models.ACMEAccount),
		events:   make(map[string]*models.CertRotationEvent),
	}
	s.seed()
	return s
}

// seed populates the store with representative mock data.
func (s *Store) seed() {
	now := time.Now().UTC()

	// --- Rotation policies ---
	policies := []*models.RotationPolicy{
		{
			ID:               "pol-le-dns",
			Name:             "Let's Encrypt / DNS-01",
			Description:      "Auto-rotate 30 days before expiry via DNS-01 challenge, deploy to Kubernetes.",
			DaysBeforeExpiry: 30,
			Provider:         models.ProviderLetsEncrypt,
			ChallengeType:    models.ChallengeDNS01,
			DeployTargets:    []models.DeployTarget{models.DeployK8sSecret},
			AutoApprove:      true,
			NotifyEmail:      "platform-eng@example.com",
			CreatedAt:        now.Add(-90 * 24 * time.Hour),
			UpdatedAt:        now.Add(-5 * 24 * time.Hour),
		},
		{
			ID:               "pol-digicert-http",
			Name:             "DigiCert / HTTP-01",
			Description:      "Enterprise cert rotation 45 days before expiry via HTTP-01, deploy to AWS ACM.",
			DaysBeforeExpiry: 45,
			Provider:         models.ProviderDigiCert,
			ChallengeType:    models.ChallengeHTTP01,
			DeployTargets:    []models.DeployTarget{models.DeployAWSACM},
			AutoApprove:      false,
			NotifyEmail:      "security@example.com",
			CreatedAt:        now.Add(-60 * 24 * time.Hour),
			UpdatedAt:        now.Add(-2 * 24 * time.Hour),
		},
		{
			ID:               "pol-internal-vault",
			Name:             "Internal CA / Azure Key Vault",
			Description:      "Internal EJBCA rotation 60 days before expiry, deploy to Azure Key Vault.",
			DaysBeforeExpiry: 60,
			Provider:         models.ProviderInternal,
			ChallengeType:    models.ChallengeDNS01,
			DeployTargets:    []models.DeployTarget{models.DeployAzureKeyVault},
			AutoApprove:      true,
			NotifyEmail:      "infra@example.com",
			CreatedAt:        now.Add(-120 * 24 * time.Hour),
			UpdatedAt:        now.Add(-10 * 24 * time.Hour),
		},
	}
	for _, p := range policies {
		s.policies[p.ID] = p
	}

	// --- ACME accounts ---
	accounts := []*models.ACMEAccount{
		{
			ID:           "acme-le-prod",
			Email:        "platform-eng@example.com",
			Provider:     models.ProviderLetsEncrypt,
			DirectoryURL: "https://acme-v02.api.letsencrypt.org/directory",
			Status:       models.ACMEAccountActive,
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
		},
		{
			ID:           "acme-le-staging",
			Email:        "platform-eng@example.com",
			Provider:     models.ProviderLetsEncrypt,
			DirectoryURL: "https://acme-staging-v02.api.letsencrypt.org/directory",
			Status:       models.ACMEAccountActive,
			CreatedAt:    now.Add(-180 * 24 * time.Hour),
		},
	}
	for _, a := range accounts {
		s.accounts[a.ID] = a
	}

	// --- Certificates at various expiry stages ---
	rotatedAt := now.Add(-45 * 24 * time.Hour)
	certs := []*models.Certificate{
		{
			ID:              "cert-001",
			Domain:          "api.example.com",
			SANs:            []string{"api-internal.example.com"},
			Issuer:          "Let's Encrypt Authority X3",
			Provider:        models.ProviderLetsEncrypt,
			IssuedAt:        now.Add(-60 * 24 * time.Hour),
			ExpiresAt:       now.Add(30 * 24 * time.Hour),
			ValidityDays:    90,
			DaysUntilExpiry: 30,
			Status:          models.CertStatusExpiring,
			AutoRenew:       true,
			RotationPolicy:  "pol-le-dns",
			DeployTargets:   []models.DeployTarget{models.DeployK8sSecret},
			Fingerprint:     "A1:B2:C3:D4:E5:F6:11:22:33:44:55:66:77:88:99:AA",
			SerialNumber:    "03:e8:01",
			LastRotatedAt:   &rotatedAt,
			DiscoveredBy:    "discovery-engine",
		},
		{
			ID:              "cert-002",
			Domain:          "app.example.com",
			SANs:            []string{"www.example.com", "example.com"},
			Issuer:          "Let's Encrypt Authority X3",
			Provider:        models.ProviderLetsEncrypt,
			IssuedAt:        now.Add(-30 * 24 * time.Hour),
			ExpiresAt:       now.Add(60 * 24 * time.Hour),
			ValidityDays:    90,
			DaysUntilExpiry: 60,
			Status:          models.CertStatusActive,
			AutoRenew:       true,
			RotationPolicy:  "pol-le-dns",
			DeployTargets:   []models.DeployTarget{models.DeployK8sSecret},
			Fingerprint:     "B2:C3:D4:E5:F6:A1:22:33:44:55:66:77:88:99:AA:BB",
			SerialNumber:    "03:e8:02",
			DiscoveredBy:    "discovery-engine",
		},
		{
			ID:              "cert-003",
			Domain:          "payments.example.com",
			SANs:            []string{"checkout.example.com"},
			Issuer:          "DigiCert Global CA G2",
			Provider:        models.ProviderDigiCert,
			IssuedAt:        now.Add(-155 * 24 * time.Hour),
			ExpiresAt:       now.Add(5 * 24 * time.Hour),
			ValidityDays:    398,
			DaysUntilExpiry: 5,
			Status:          models.CertStatusExpiring,
			AutoRenew:       true,
			RotationPolicy:  "pol-digicert-http",
			DeployTargets:   []models.DeployTarget{models.DeployAWSACM},
			Fingerprint:     "C3:D4:E5:F6:A1:B2:33:44:55:66:77:88:99:AA:BB:CC",
			SerialNumber:    "07:d0:03",
			DiscoveredBy:    "discovery-engine",
		},
		{
			ID:              "cert-004",
			Domain:          "legacy-erp.corp.example.com",
			Issuer:          "Internal EJBCA Root CA",
			Provider:        models.ProviderInternal,
			IssuedAt:        now.Add(-380 * 24 * time.Hour),
			ExpiresAt:       now.Add(-20 * 24 * time.Hour),
			ValidityDays:    398,
			DaysUntilExpiry: -20,
			Status:          models.CertStatusExpired,
			AutoRenew:       false,
			RotationPolicy:  "pol-internal-vault",
			DeployTargets:   []models.DeployTarget{models.DeployAzureKeyVault},
			Fingerprint:     "D4:E5:F6:A1:B2:C3:44:55:66:77:88:99:AA:BB:CC:DD",
			SerialNumber:    "0b:b8:04",
			DiscoveredBy:    "discovery-engine",
		},
		{
			ID:              "cert-005",
			Domain:          "cdn.example.com",
			SANs:            []string{"assets.example.com", "static.example.com"},
			Issuer:          "Let's Encrypt Authority X3",
			Provider:        models.ProviderLetsEncrypt,
			IssuedAt:        now.Add(-5 * 24 * time.Hour),
			ExpiresAt:       now.Add(85 * 24 * time.Hour),
			ValidityDays:    90,
			DaysUntilExpiry: 85,
			Status:          models.CertStatusActive,
			AutoRenew:       true,
			RotationPolicy:  "pol-le-dns",
			DeployTargets:   []models.DeployTarget{models.DeployK8sSecret, models.DeployAWSACM},
			Fingerprint:     "E5:F6:A1:B2:C3:D4:55:66:77:88:99:AA:BB:CC:DD:EE",
			SerialNumber:    "0f:a0:05",
			DiscoveredBy:    "discovery-engine",
		},
		{
			ID:              "cert-006",
			Domain:          "auth.example.com",
			Issuer:          "Sectigo RSA Domain Validation Secure Server CA",
			Provider:        models.ProviderSectigo,
			IssuedAt:        now.Add(-10 * 24 * time.Hour),
			ExpiresAt:       now.Add(80 * 24 * time.Hour),
			ValidityDays:    90,
			DaysUntilExpiry: 80,
			Status:          models.CertStatusRenewing,
			AutoRenew:       true,
			DeployTargets:   []models.DeployTarget{models.DeployK8sSecret},
			Fingerprint:     "F6:A1:B2:C3:D4:E5:66:77:88:99:AA:BB:CC:DD:EE:FF",
			SerialNumber:    "13:88:06",
			DiscoveredBy:    "discovery-engine",
		},
	}
	for _, c := range certs {
		s.certs[c.ID] = c
	}

	// Seed a rotation event for cert-003 (the nearly-expired payments cert).
	eventTime := now.Add(-1 * time.Hour)
	s.events["evt-001"] = &models.CertRotationEvent{
		ID:          "evt-001",
		CertID:      "cert-003",
		Domain:      "payments.example.com",
		Status:      models.RotationEventPending,
		TriggeredBy: "policy-engine",
		StartedAt:   eventTime,
	}
}

// ListCertificates returns all certificates in the inventory.
func (s *Store) ListCertificates() []*models.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.Certificate, 0, len(s.certs))
	for _, c := range s.certs {
		out = append(out, c)
	}
	return out
}

// GetCertificate returns a single certificate by ID.
func (s *Store) GetCertificate(id string) (*models.Certificate, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.certs[id]
	return c, ok
}

// AddCertificate inserts a new certificate record into the inventory.
func (s *Store) AddCertificate(cert *models.Certificate) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.certs[cert.ID] = cert
}

// ListPolicies returns all rotation policies.
func (s *Store) ListPolicies() []*models.RotationPolicy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.RotationPolicy, 0, len(s.policies))
	for _, p := range s.policies {
		out = append(out, p)
	}
	return out
}

// GetPolicy returns a single rotation policy by ID.
func (s *Store) GetPolicy(id string) (*models.RotationPolicy, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.policies[id]
	return p, ok
}

// AddPolicy inserts a new rotation policy.
func (s *Store) AddPolicy(policy *models.RotationPolicy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.policies[policy.ID] = policy
}

// ListACMEAccounts returns all registered ACME accounts.
func (s *Store) ListACMEAccounts() []*models.ACMEAccount {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]*models.ACMEAccount, 0, len(s.accounts))
	for _, a := range s.accounts {
		out = append(out, a)
	}
	return out
}

// AddRotationEvent records a new rotation attempt.
func (s *Store) AddRotationEvent(event *models.CertRotationEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events[event.ID] = event
}

// ListRotationEvents returns all events for a given certificate.
func (s *Store) ListRotationEvents(certID string) []*models.CertRotationEvent {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []*models.CertRotationEvent
	for _, e := range s.events {
		if e.CertID == certID {
			out = append(out, e)
		}
	}
	return out
}

// UpdateCertStatus updates the Status field of an existing certificate.
func (s *Store) UpdateCertStatus(id string, status models.CertStatus) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.certs[id]
	if !ok {
		return false
	}
	c.Status = status
	return true
}

// nextID generates a simple sequential-style ID with a given prefix.
func nextID(prefix string, n int) string {
	return fmt.Sprintf("%s-%03d", prefix, n)
}

// NextCertID returns the next available certificate ID.
func (s *Store) NextCertID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return nextID("cert", len(s.certs)+1)
}

// NextPolicyID returns the next available policy ID.
func (s *Store) NextPolicyID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return nextID("pol-custom", len(s.policies)+1)
}

// NextEventID returns the next available event ID.
func (s *Store) NextEventID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return nextID("evt", len(s.events)+1)
}
