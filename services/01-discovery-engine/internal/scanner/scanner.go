// Package scanner implements the core asset and certificate discovery logic.
// In this phase the cloud/on-prem probes return deterministic stub data so the
// service compiles and runs without live cloud credentials.  Real provider SDK
// calls are wired in subsequent phases behind the same Scanner interface.
package scanner

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/ncsound919/modernization-control-plane/services/discovery-engine/internal/models"
)

// Config holds runtime configuration sourced from environment variables.
type Config struct {
	Neo4jURI      string
	Neo4jUser     string
	Neo4jPassword string
	PostgresDSN   string
	Port          string
}

// Scanner is the top-level discovery orchestrator.
type Scanner struct {
	cfg    Config
	mu     sync.RWMutex
	scans  map[string]*models.Scan
	assets []models.Asset
	certs  []models.Certificate
}

// New constructs a Scanner from the supplied configuration.
func New(cfg Config) *Scanner {
	s := &Scanner{
		cfg:   cfg,
		scans: make(map[string]*models.Scan),
	}
	// Pre-populate with stub discovered assets so the API returns data immediately.
	s.assets = stubAssets()
	s.certs = stubCertificates()
	return s
}

// StartScan begins an asynchronous scan and returns the Scan record immediately.
func (s *Scanner) StartScan(ctx context.Context, req models.ScanRequest) *models.Scan {
	scan := &models.Scan{
		ID:        newID("scan"),
		Status:    models.ScanStatusPending,
		Request:   req,
		StartTime: time.Now().UTC(),
	}
	s.mu.Lock()
	s.scans[scan.ID] = scan
	s.mu.Unlock()

	go s.runScan(scan)
	return scan
}

// GetScan retrieves a scan by ID.
func (s *Scanner) GetScan(id string) (*models.Scan, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sc, ok := s.scans[id]
	return sc, ok
}

// Assets returns the current asset list.
func (s *Scanner) Assets() []models.Asset {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Asset, len(s.assets))
	copy(out, s.assets)
	return out
}

// Certificates returns the current certificate list.
func (s *Scanner) Certificates() []models.Certificate {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Certificate, len(s.certs))
	copy(out, s.certs)
	return out
}

// BuildGraph constructs a DiscoveryGraph from all known assets and certificates.
func (s *Scanner) BuildGraph(scanID string) models.DiscoveryGraph {
	s.mu.RLock()
	defer s.mu.RUnlock()

	graph := models.DiscoveryGraph{
		ScanID:      scanID,
		GeneratedAt: time.Now().UTC(),
	}

	for _, a := range s.assets {
		graph.Nodes = append(graph.Nodes, models.GraphNode{
			ID:    a.ID,
			Label: string(a.Type),
			Name:  a.Name,
			Props: map[string]interface{}{
				"environment": string(a.Environment),
				"risk_score":  a.RiskScore,
				"region":      a.Region,
			},
		})
	}

	for _, c := range s.certs {
		graph.Nodes = append(graph.Nodes, models.GraphNode{
			ID:    c.ID,
			Label: "certificate",
			Name:  c.Domain,
			Props: map[string]interface{}{
				"days_to_expiry": c.DaysToExpiry,
				"is_expired":     c.IsExpired,
				"risk_score":     c.RiskScore,
			},
		})
	}

	// Wire stub edges: services depend on databases, certs secure services.
	for i, a := range s.assets {
		if a.Type == models.AssetTypeService && i < len(s.assets)-1 {
			db := s.findFirstAsset(models.AssetTypeDatabase)
			if db != nil {
				graph.Edges = append(graph.Edges, models.GraphEdge{
					Source: a.ID,
					Target: db.ID,
					Type:   "depends_on",
					Weight: 1.0,
				})
			}
		}
		if a.Type == models.AssetTypeService && len(s.certs) > 0 {
			graph.Edges = append(graph.Edges, models.GraphEdge{
				Source: a.ID,
				Target: s.certs[0].ID,
				Type:   "secured_by",
				Weight: 1.0,
			})
		}
	}

	return graph
}

// runScan simulates a real scan with a short delay then updates results.
func (s *Scanner) runScan(scan *models.Scan) {
	s.mu.Lock()
	scan.Status = models.ScanStatusRunning
	s.mu.Unlock()

	// Probe real TLS targets if supplied; fall back to stubs.
	var certResults []models.Certificate
	for _, target := range scan.Request.Targets {
		if cert, err := probeTLS(target); err == nil {
			certResults = append(certResults, cert)
		}
	}

	// Simulate work for stub targets.
	time.Sleep(500 * time.Millisecond)

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(certResults) > 0 {
		s.certs = append(s.certs, certResults...)
	}

	now := time.Now().UTC()
	scan.EndTime = &now
	scan.Status = models.ScanStatusCompleted
	scan.Results = &models.ScanResult{
		AssetsFound:       len(s.assets),
		CertificatesFound: len(s.certs),
		LegacyFound:       countLegacy(s.assets),
		RiskyAssets:       countRisky(s.assets),
	}
}

// probeTLS performs a real TLS handshake against host:443 and extracts cert info.
func probeTLS(target string) (models.Certificate, error) {
	host, port, err := net.SplitHostPort(target)
	if err != nil {
		host = target
		port = "443"
	}

	dialer := &net.Dialer{Timeout: 5 * time.Second}
	conn, err := tls.DialWithDialer(dialer, "tcp", net.JoinHostPort(host, port), &tls.Config{})
	if err != nil {
		return models.Certificate{}, fmt.Errorf("tls dial %s: %w", target, err)
	}
	defer conn.Close()

	chains := conn.ConnectionState().PeerCertificates
	if len(chains) == 0 {
		return models.Certificate{}, fmt.Errorf("no peer certificates from %s", target)
	}

	leaf := chains[0]
	now := time.Now().UTC()
	daysLeft := int(leaf.NotAfter.Sub(now).Hours() / 24)

	var sans []string
	sans = append(sans, leaf.DNSNames...)
	for _, ip := range leaf.IPAddresses {
		sans = append(sans, ip.String())
	}

	return models.Certificate{
		ID:             newID("cert"),
		Domain:         host,
		Issuer:         leaf.Issuer.CommonName,
		Subject:        leaf.Subject.CommonName,
		NotBefore:      leaf.NotBefore.UTC(),
		NotAfter:       leaf.NotAfter.UTC(),
		DaysToExpiry:   daysLeft,
		IsExpired:      daysLeft < 0,
		IsExpiringSoon: daysLeft >= 0 && daysLeft <= 30,
		SANs:           sans,
		RiskScore:      certRiskScore(daysLeft),
		DiscoveredAt:   now,
	}, nil
}

// certRiskScore maps certificate days-to-expiry to a 0–10 risk score.
func certRiskScore(daysLeft int) float64 {
	switch {
	case daysLeft < 0:
		return 10.0
	case daysLeft <= 7:
		return 9.0
	case daysLeft <= 14:
		return 7.5
	case daysLeft <= 30:
		return 5.0
	case daysLeft <= 90:
		return 2.0
	default:
		return 0.5
	}
}

// findFirstAsset returns the first asset of the given type (must hold read lock).
func (s *Scanner) findFirstAsset(t models.AssetType) *models.Asset {
	for i := range s.assets {
		if s.assets[i].Type == t {
			return &s.assets[i]
		}
	}
	return nil
}

func countLegacy(assets []models.Asset) int {
	n := 0
	for _, a := range assets {
		if a.Type == models.AssetTypeLegacy {
			n++
		}
	}
	return n
}

func countRisky(assets []models.Asset) int {
	n := 0
	for _, a := range assets {
		if a.RiskScore >= 7.0 {
			n++
		}
	}
	return n
}

// newID generates a cryptographically random unique identifier with a prefix.
func newID(prefix string) string {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		// Fall back to timestamp-only ID if crypto/rand is unavailable.
		return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
	}
	return fmt.Sprintf("%s-%d-%s", prefix, time.Now().UnixNano(), hex.EncodeToString(b))
}

// stubAssets returns representative stub assets for Phase 1.
func stubAssets() []models.Asset {
	now := time.Now().UTC()
	return []models.Asset{
		{
			ID: "asset-001", Name: "payments-api", Type: models.AssetTypeService,
			Environment: models.EnvAWS, Region: "us-east-1", RiskScore: 3.2,
			Tags: map[string]string{"team": "platform", "tier": "prod"},
			DiscoveredAt: now,
		},
		{
			ID: "asset-002", Name: "claims-db-oracle", Type: models.AssetTypeDatabase,
			Environment: models.EnvOnPrem, Region: "dc-primary", RiskScore: 8.7,
			Tags: map[string]string{"engine": "oracle-11g", "tier": "prod"},
			Metadata: map[string]string{"version": "11.2.0.4", "eol": "2020-12-31"},
			DiscoveredAt: now,
		},
		{
			ID: "asset-003", Name: "cobol-batch-processor", Type: models.AssetTypeLegacy,
			Environment: models.EnvOnPrem, Region: "mainframe-z16", RiskScore: 9.5,
			Tags: map[string]string{"language": "COBOL", "system": "mainframe"},
			Metadata: map[string]string{"system_type": "COBOL", "confidence": "0.97"},
			DiscoveredAt: now,
		},
		{
			ID: "asset-004", Name: "hl7-gateway", Type: models.AssetTypeService,
			Environment: models.EnvAzure, Region: "eastus2", RiskScore: 6.1,
			Tags: map[string]string{"protocol": "HL7v2", "team": "health"},
			Metadata: map[string]string{"system_type": "HL7", "version": "2.5"},
			DiscoveredAt: now,
		},
		{
			ID: "asset-005", Name: "vsam-inventory-store", Type: models.AssetTypeLegacy,
			Environment: models.EnvOnPrem, Region: "mainframe-z16", RiskScore: 8.9,
			Tags: map[string]string{"storage": "VSAM", "system": "mainframe"},
			Metadata: map[string]string{"system_type": "VSAM", "confidence": "0.91"},
			DiscoveredAt: now,
		},
		{
			ID: "asset-006", Name: "notification-queue", Type: models.AssetTypeQueue,
			Environment: models.EnvAWS, Region: "us-west-2", RiskScore: 1.5,
			Tags: map[string]string{"team": "platform"},
			DiscoveredAt: now,
		},
		{
			ID: "asset-007", Name: "fhir-r4-server", Type: models.AssetTypeService,
			Environment: models.EnvGCP, Region: "us-central1", RiskScore: 2.8,
			Tags: map[string]string{"protocol": "FHIR-R4", "team": "health"},
			DiscoveredAt: now,
		},
	}
}

// stubCertificates returns representative stub certificates for Phase 1.
func stubCertificates() []models.Certificate {
	now := time.Now().UTC()
	return []models.Certificate{
		{
			ID: "cert-001", Domain: "payments.example.internal",
			Issuer: "Example Internal CA", Subject: "payments.example.internal",
			NotBefore: now.AddDate(0, -6, 0), NotAfter: now.AddDate(0, 6, 0),
			DaysToExpiry: 180, IsExpired: false, IsExpiringSoon: false,
			SANs: []string{"payments.example.internal", "pay.example.internal"},
			RiskScore: 0.5, DiscoveredAt: now,
		},
		{
			ID: "cert-002", Domain: "legacy-claims.example.internal",
			Issuer: "Example Internal CA", Subject: "legacy-claims.example.internal",
			NotBefore: now.AddDate(-1, 0, 0), NotAfter: now.AddDate(0, 0, 12),
			DaysToExpiry: 12, IsExpired: false, IsExpiringSoon: true,
			RiskScore: 7.5, DiscoveredAt: now,
		},
		{
			ID: "cert-003", Domain: "hl7-gw.example.internal",
			Issuer: "DigiCert Inc", Subject: "hl7-gw.example.internal",
			NotBefore: now.AddDate(0, -3, 0), NotAfter: now.AddDate(0, 9, 0),
			DaysToExpiry: 270, IsExpired: false, IsExpiringSoon: false,
			RiskScore: 0.5, DiscoveredAt: now,
		},
		{
			ID: "cert-004", Domain: "old-portal.example.internal",
			Issuer: "Legacy Corp CA", Subject: "old-portal.example.internal",
			NotBefore: now.AddDate(-2, 0, 0), NotAfter: now.AddDate(0, 0, -5),
			DaysToExpiry: -5, IsExpired: true, IsExpiringSoon: false,
			RiskScore: 10.0, DiscoveredAt: now,
		},
	}
}
