# Module 07 — Governance Engine

## Purpose
Provides real-time policy enforcement, risk controls, and hard kill switches for all modernization operations. The reason CISOs and compliance officers approve the platform. With 40% of AI agent projects predicted to fail due to policy violations, governance is a first-class concern.

## Capabilities

### Policy Engine (OPA-Based)
- Policies defined as code (Rego language via Open Policy Agent)
- Built-in policy packs for HIPAA, GDPR, SOC 2, PCI-DSS
- Custom policies per tenant or vertical
- Inline enforcement at every high-risk operation (agent actions, data writes, cert operations)

### Example Policies
```rego
# No agent can access PHI data without human approval
deny[msg] {
    input.action == "read"
    input.data_classification == "PHI"
    not input.human_approved
    msg := "PHI access requires explicit human approval"
}

# Workflow cost cap
deny[msg] {
    input.workflow_cost_usd > data.tenant.cost_limit_usd
    msg := sprintf("Cost limit exceeded: $%v > $%v", [input.workflow_cost_usd, data.tenant.cost_limit_usd])
}
```

### Kill Switches
- **Tenant-level:** Immediately stops all agents and write-operations for a tenant
- **Workflow-level:** Pauses a specific pipeline on threshold breach
- **Emergency mode:** Read-only lockdown with notification to on-call

### Audit Trail
- Immutable append-only log of every agent action, prompt/result, and config change
- Tamper-evident with cryptographic chaining
- Export to SIEM/GRC systems (Splunk, Elastic, ServiceNow)
- Retention policies per compliance framework

## Tech Stack
- **Language:** Go 1.22+
- **Policy Engine:** Open Policy Agent (OPA) + Rego
- **Audit Store:** PostgreSQL (append-only) + write-ahead log
- **Event Stream:** Apache Kafka
- **Alerting:** PagerDuty, OpsGenie, Slack integrations

## Directory Structure
```
07-governance-engine/
├── cmd/
│   └── governance/       # Main governance service
├── internal/
│   ├── policy/           # OPA integration and policy evaluator
│   ├── killswitch/       # Kill switch mechanisms
│   ├── audit/            # Immutable audit log writer
│   └── alerts/           # Alert routing and integrations
├── policies/
│   ├── hipaa.rego
│   ├── gdpr.rego
│   ├── soc2.rego
│   └── custom/           # Tenant-specific policies
├── Dockerfile
├── go.mod
└── README.md
```

## Key Metrics
- Policy violations caught vs. escaped
- Time to investigate and explain a given change (audit query time)
- % of agent workflows operating within defined policy bounds

## Status
`Phase 2-4 — Parallel development with agent orchestrator`
