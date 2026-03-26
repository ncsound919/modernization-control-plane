// Package models defines the core data structures for the Certificate Lifecycle Manager.
package models

import "time"

// CertStatus represents the lifecycle state of a managed certificate.
type CertStatus string

const (
	CertStatusActive   CertStatus = "active"
	CertStatusExpiring CertStatus = "expiring_soon"
	CertStatusExpired  CertStatus = "expired"
	CertStatusRenewing CertStatus = "renewing"
	CertStatusRevoked  CertStatus = "revoked"
)

// ChallengeType is the ACME challenge mechanism used to prove domain ownership.
type ChallengeType string

const (
	ChallengeDNS01  ChallengeType = "dns-01"
	ChallengeHTTP01 ChallengeType = "http-01"
)

// Provider is the certificate authority or deployment target.
type Provider string

const (
	ProviderLetsEncrypt Provider = "letsencrypt"
	ProviderDigiCert    Provider = "digicert"
	ProviderSectigo     Provider = "sectigo"
	ProviderInternal    Provider = "internal-ejbca"
)

// DeployTarget identifies where renewed certificates are deployed.
type DeployTarget string

const (
	DeployK8sSecret    DeployTarget = "kubernetes-secret"
	DeployAWSACM       DeployTarget = "aws-acm"
	DeployAzureKeyVault DeployTarget = "azure-key-vault"
)

// ACMEAccountStatus reflects the registration state of an ACME account.
type ACMEAccountStatus string

const (
	ACMEAccountActive  ACMEAccountStatus = "active"
	ACMEAccountPending ACMEAccountStatus = "pending"
)

// RotationEventStatus is the outcome of a single rotation attempt.
type RotationEventStatus string

const (
	RotationEventSuccess RotationEventStatus = "success"
	RotationEventFailed  RotationEventStatus = "failed"
	RotationEventPending RotationEventStatus = "pending"
)

// Certificate represents a TLS/SSL certificate tracked in the CLM inventory.
type Certificate struct {
	ID             string        `json:"id"`
	Domain         string        `json:"domain"`
	SANs           []string      `json:"sans,omitempty"`
	Issuer         string        `json:"issuer"`
	Provider       Provider      `json:"provider"`
	IssuedAt       time.Time     `json:"issued_at"`
	ExpiresAt      time.Time     `json:"expires_at"`
	ValidityDays   int           `json:"validity_days"`
	DaysUntilExpiry int          `json:"days_until_expiry"`
	Status         CertStatus    `json:"status"`
	AutoRenew      bool          `json:"auto_renew"`
	RotationPolicy string        `json:"rotation_policy_id,omitempty"`
	DeployTargets  []DeployTarget `json:"deploy_targets,omitempty"`
	Fingerprint    string        `json:"fingerprint"`
	SerialNumber   string        `json:"serial_number"`
	LastRotatedAt  *time.Time    `json:"last_rotated_at,omitempty"`
	DiscoveredBy   string        `json:"discovered_by,omitempty"`
}

// RotationPolicy defines when and how a certificate should be rotated.
type RotationPolicy struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	DaysBeforeExpiry int           `json:"days_before_expiry"`
	Provider         Provider      `json:"provider"`
	ChallengeType    ChallengeType `json:"challenge_type"`
	DeployTargets    []DeployTarget `json:"deploy_targets"`
	AutoApprove      bool          `json:"auto_approve"`
	NotifyEmail      string        `json:"notify_email,omitempty"`
	CreatedAt        time.Time     `json:"created_at"`
	UpdatedAt        time.Time     `json:"updated_at"`
}

// ACMEAccount represents a registered account with an ACME-compatible CA.
type ACMEAccount struct {
	ID           string            `json:"id"`
	Email        string            `json:"email"`
	Provider     Provider          `json:"provider"`
	DirectoryURL string            `json:"directory_url"`
	Status       ACMEAccountStatus `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
}

// CertRotationEvent records a single certificate rotation attempt.
type CertRotationEvent struct {
	ID          string              `json:"id"`
	CertID      string              `json:"cert_id"`
	Domain      string              `json:"domain"`
	Status      RotationEventStatus `json:"status"`
	TriggeredBy string              `json:"triggered_by"`
	StartedAt   time.Time           `json:"started_at"`
	CompletedAt *time.Time          `json:"completed_at,omitempty"`
	ErrorMsg    string              `json:"error_msg,omitempty"`
	NewExpiry   *time.Time          `json:"new_expiry,omitempty"`
}

// AddCertRequest is the payload for adding a certificate to the inventory.
type AddCertRequest struct {
	Domain        string        `json:"domain"`
	SANs          []string      `json:"sans,omitempty"`
	Provider      Provider      `json:"provider"`
	ChallengeType ChallengeType `json:"challenge_type"`
	AutoRenew     bool          `json:"auto_renew"`
	PolicyID      string        `json:"rotation_policy_id,omitempty"`
	DeployTargets []DeployTarget `json:"deploy_targets,omitempty"`
}

// CreatePolicyRequest is the payload for creating a new rotation policy.
type CreatePolicyRequest struct {
	Name             string        `json:"name"`
	Description      string        `json:"description"`
	DaysBeforeExpiry int           `json:"days_before_expiry"`
	Provider         Provider      `json:"provider"`
	ChallengeType    ChallengeType `json:"challenge_type"`
	DeployTargets    []DeployTarget `json:"deploy_targets"`
	AutoApprove      bool          `json:"auto_approve"`
	NotifyEmail      string        `json:"notify_email,omitempty"`
}
