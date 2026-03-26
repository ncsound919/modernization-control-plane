// Package acme provides a stub ACME v2 client for Let's Encrypt and enterprise CAs.
//
// In production this wraps golang.org/x/crypto/acme to perform real certificate
// issuance via DNS-01 or HTTP-01 challenges. The stub is provided so the service
// starts and serves meaningful responses without requiring an external CA or
// network connectivity.
package acme

import (
	"log/slog"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/clm-service/internal/models"
)

// Client is a stub ACME v2 client.
type Client struct {
	logger *slog.Logger
}

// New creates an ACME client. The directoryURL would normally point to the ACME
// CA directory (e.g. https://acme-v02.api.letsencrypt.org/directory).
func New(logger *slog.Logger) *Client {
	return &Client{logger: logger}
}

// IssueResult holds the outcome of a certificate issuance attempt.
type IssueResult struct {
	Domain      string
	IssuedAt    time.Time
	ExpiresAt   time.Time
	Fingerprint string
	SerialNumber string
	PEMCert     string
}

// Issue performs a stub certificate issuance for the given domain.
//
// A real implementation would:
//  1. Register or load an ACME account key from Vault.
//  2. Create an ACME order for the domain + SANs.
//  3. Resolve the DNS-01 or HTTP-01 challenge.
//  4. Poll until the CA marks the challenge valid.
//  5. Finalize the order and download the signed certificate.
//  6. Store the private key in HashiCorp Vault.
//  7. Deploy the cert to the configured DeployTargets (K8s Secret, ACM, Key Vault).
func (c *Client) Issue(domain string, challenge models.ChallengeType, provider models.Provider) (*IssueResult, error) {
	c.logger.Info("acme: issuing certificate (stub)",
		"domain", domain,
		"challenge", challenge,
		"provider", provider,
	)

	now := time.Now().UTC()
	result := &IssueResult{
		Domain:       domain,
		IssuedAt:     now,
		ExpiresAt:    now.Add(90 * 24 * time.Hour),
		Fingerprint:  "00:11:22:33:44:55:66:77:88:99:AA:BB:CC:DD:EE:FF",
		SerialNumber: "de:ad:be:ef",
		PEMCert:      "-----BEGIN CERTIFICATE-----\n(stub)\n-----END CERTIFICATE-----\n",
	}

	c.logger.Info("acme: certificate issued (stub)",
		"domain", domain,
		"expires_at", result.ExpiresAt,
	)
	return result, nil
}

// RenewChallengeDNS01 logs the DNS TXT record that would be set for a DNS-01 challenge.
func (c *Client) RenewChallengeDNS01(domain, token string) {
	c.logger.Info("acme: DNS-01 challenge",
		"record", "_acme-challenge."+domain,
		"value", token,
	)
}

// RenewChallengeHTTP01 logs the HTTP path that would serve the HTTP-01 challenge token.
func (c *Client) RenewChallengeHTTP01(domain, token, keyAuth string) {
	c.logger.Info("acme: HTTP-01 challenge",
		"path", "/.well-known/acme-challenge/"+token,
		"domain", domain,
	)
}
