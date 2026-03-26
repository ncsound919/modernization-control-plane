package models

import "time"

type PolicyFramework string

const (
	FrameworkHIPAA  PolicyFramework = "HIPAA"
	FrameworkGDPR   PolicyFramework = "GDPR"
	FrameworkSOC2   PolicyFramework = "SOC2"
	FrameworkPCIDSS PolicyFramework = "PCI-DSS"
	FrameworkCustom PolicyFramework = "CUSTOM"
)

type Policy struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Type        string          `json:"type"`
	Framework   PolicyFramework `json:"framework"`
	Rego        string          `json:"rego"`
	Enabled     bool            `json:"enabled"`
	TenantID    string          `json:"tenant_id,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type PolicyEvaluationRequest struct {
	PolicyID string                 `json:"policy_id,omitempty"`
	Framework PolicyFramework       `json:"framework,omitempty"`
	Input    map[string]interface{} `json:"input"`
	TenantID string                 `json:"tenant_id,omitempty"`
}

type PolicyDecision struct {
	Allowed   bool            `json:"allowed"`
	Reason    string          `json:"reason"`
	Violations []string       `json:"violations,omitempty"`
	Policy    *Policy         `json:"policy,omitempty"`
	Framework PolicyFramework `json:"framework,omitempty"`
	EvaluatedAt time.Time     `json:"evaluated_at"`
}

type KillSwitchScope string

const (
	ScopeTenant    KillSwitchScope = "tenant"
	ScopeWorkflow  KillSwitchScope = "workflow"
	ScopeEmergency KillSwitchScope = "emergency"
	ScopeGlobal    KillSwitchScope = "global"
)

type KillSwitch struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Active      bool            `json:"active"`
	Scope       KillSwitchScope `json:"scope"`
	TenantID    string          `json:"tenant_id,omitempty"`
	WorkflowID  string          `json:"workflow_id,omitempty"`
	ActivatedBy string          `json:"activated_by,omitempty"`
	ActivatedAt *time.Time      `json:"activated_at,omitempty"`
	Reason      string          `json:"reason,omitempty"`
}

type AuditEntry struct {
	ID         string                 `json:"id"`
	Timestamp  time.Time              `json:"timestamp"`
	Actor      string                 `json:"actor"`
	Action     string                 `json:"action"`
	Resource   string                 `json:"resource"`
	TenantID   string                 `json:"tenant_id,omitempty"`
	WorkflowID string                 `json:"workflow_id,omitempty"`
	Decision   string                 `json:"decision"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Hash       string                 `json:"hash"`
	PrevHash   string                 `json:"prev_hash"`
}

type ComplianceViolation struct {
	RuleID      string    `json:"rule_id"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	DetectedAt  time.Time `json:"detected_at"`
}

type ComplianceStatus struct {
	Framework   PolicyFramework       `json:"framework"`
	Status      string                `json:"status"`
	Score       int                   `json:"score"`
	Violations  []ComplianceViolation `json:"violations"`
	LastChecked time.Time             `json:"last_checked"`
}
