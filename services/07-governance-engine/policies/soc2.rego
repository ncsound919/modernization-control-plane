package soc2

# SOC 2 Trust Services Criteria enforcement policies.
# Covers the five Trust Services Categories: Security (CC), Availability (A),
# Processing Integrity (PI), Confidentiality (C), and Privacy (P).

default allow = false

allow {
    count(deny) == 0
}

# --- CC6: Logical and Physical Access Controls ---

# Least privilege: guest accounts may not perform write or admin actions.
deny[msg] {
    lower(input.actor_role) == "guest"
    privileged_action
    msg := "SOC 2 CC6.3: guest role may not perform privileged operations (write/delete/admin)"
}

privileged_action {
    action := lower(input.action)
    action == "write"
}
privileged_action {
    action := lower(input.action)
    action == "delete"
}
privileged_action {
    action := lower(input.action)
    action == "admin"
}
privileged_action {
    action := lower(input.action)
    action == "write_sensitive"
}

# MFA required for privileged operations.
deny[msg] {
    privileged_action
    not input.mfa_verified
    msg := "SOC 2 CC6.1: multi-factor authentication is required for privileged operations"
}

# --- CC7: System Operations ---

# Unauthenticated requests to non-public resources must be denied.
deny[msg] {
    not input.is_public_resource
    not input.actor_authenticated
    msg := "SOC 2 CC7.1: unauthenticated access to non-public resources is prohibited"
}

# --- CC8: Change Management ---

# Production changes require change-request approval.
deny[msg] {
    input.environment == "production"
    input.action == "deploy"
    not input.change_request_approved
    msg := "SOC 2 CC8.1: production deployments require an approved change request"
}

# --- A1: Availability ---

# Workflow cost caps must not be breached.
deny[msg] {
    input.workflow_cost_usd > input.tenant_cost_limit_usd
    msg := sprintf("SOC 2 A1.1: workflow cost cap exceeded ($%.2f > $%.2f)", [input.workflow_cost_usd, input.tenant_cost_limit_usd])
}

# --- PI1: Processing Integrity ---

# Data-altering operations must declare a schema version to prevent schema drift.
deny[msg] {
    input.action == "write"
    not input.schema_version
    msg := "SOC 2 PI1.1: write operations must declare a schema version"
}

# --- C1: Confidentiality ---

# Confidential data must not be exported without explicit authorisation.
deny[msg] {
    lower(input.data_classification) == "confidential"
    input.action == "export"
    not input.export_authorised
    msg := "SOC 2 C1.1: export of confidential data requires explicit authorisation"
}
