package models

import "time"

// AssetType categorises discovered infrastructure resources.
type AssetType string

const (
	AssetTypeService     AssetType = "service"
	AssetTypeDatabase    AssetType = "database"
	AssetTypeCertificate AssetType = "certificate"
	AssetTypeQueue       AssetType = "queue"
	AssetTypeLegacy      AssetType = "legacy"
	AssetTypeUnknown     AssetType = "unknown"
)

// Environment describes where the asset lives.
type Environment string

const (
	EnvAWS    Environment = "aws"
	EnvAzure  Environment = "azure"
	EnvGCP    Environment = "gcp"
	EnvOnPrem Environment = "on-prem"
)

// ScanStatus tracks lifecycle of a scan request.
type ScanStatus string

const (
	ScanStatusPending   ScanStatus = "pending"
	ScanStatusRunning   ScanStatus = "running"
	ScanStatusCompleted ScanStatus = "completed"
	ScanStatusFailed    ScanStatus = "failed"
)

// Asset represents any discovered infrastructure component.
type Asset struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Type        AssetType         `json:"type"`
	Environment Environment       `json:"environment"`
	Region      string            `json:"region,omitempty"`
	RiskScore   float64           `json:"risk_score"` // 0.0–10.0
	Tags        map[string]string `json:"tags,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	DiscoveredAt time.Time        `json:"discovered_at"`
}

// Certificate holds TLS/SSL certificate details.
type Certificate struct {
	ID          string    `json:"id"`
	Domain      string    `json:"domain"`
	Issuer      string    `json:"issuer"`
	Subject     string    `json:"subject"`
	NotBefore   time.Time `json:"not_before"`
	NotAfter    time.Time `json:"not_after"`
	DaysToExpiry int      `json:"days_to_expiry"`
	IsExpired   bool      `json:"is_expired"`
	IsExpiringSoon bool      `json:"is_expiring_soon"` // within 30 days
	SANs        []string  `json:"sans,omitempty"`
	RiskScore   float64   `json:"risk_score"`
	DiscoveredAt time.Time `json:"discovered_at"`
}

// LegacyFingerprint captures legacy system detection results.
type LegacyFingerprint struct {
	AssetID    string    `json:"asset_id"`
	SystemType string    `json:"system_type"` // COBOL, VSAM, DB2, HL7, FHIR, EMR, EHR
	Confidence float64   `json:"confidence"`  // 0.0–1.0
	Indicators []string  `json:"indicators,omitempty"`
	DetectedAt time.Time `json:"detected_at"`
}

// ScanRequest is the payload for triggering a new scan.
type ScanRequest struct {
	Targets     []string    `json:"targets"`
	Environment Environment `json:"environment"`
	ScanTypes   []string    `json:"scan_types"` // certs, assets, legacy
}

// ScanResult aggregates outcomes from a completed scan.
type ScanResult struct {
	AssetsFound       int `json:"assets_found"`
	CertificatesFound int `json:"certificates_found"`
	LegacyFound       int `json:"legacy_found"`
	RiskyAssets       int `json:"risky_assets"`
}

// Scan represents a single scan lifecycle.
type Scan struct {
	ID          string      `json:"id"`
	Status      ScanStatus  `json:"status"`
	Request     ScanRequest `json:"request"`
	StartTime   time.Time   `json:"start_time"`
	EndTime     *time.Time  `json:"end_time,omitempty"`
	Results     *ScanResult `json:"results,omitempty"`
	Error       string      `json:"error,omitempty"`
}

// GraphNode is a vertex in the Discovery Graph.
type GraphNode struct {
	ID    string            `json:"id"`
	Label string            `json:"label"` // app, db, cert, queue, legacy
	Name  string            `json:"name"`
	Props map[string]interface{} `json:"props,omitempty"`
}

// GraphEdge is a directed relationship between two nodes.
type GraphEdge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Type   string `json:"type"` // calls, owns, depends_on, secured_by
	Weight float64 `json:"weight,omitempty"`
}

// DiscoveryGraph is the full technical-debt graph for a scan.
type DiscoveryGraph struct {
	ScanID string      `json:"scan_id"`
	Nodes  []GraphNode `json:"nodes"`
	Edges  []GraphEdge `json:"edges"`
	GeneratedAt time.Time `json:"generated_at"`
}
