package hipaa

# HIPAA Security Rule and Privacy Rule enforcement policies.
# These Rego rules are evaluated by the OPA engine inside the governance service.

default allow = false

# Allow the request only when no deny rules fire.
allow {
    count(deny) == 0
}

# --- Privacy Rule ---

# PHI access requires explicit human approval.
deny[msg] {
    input.action == "read"
    lower(input.data_classification) == "phi"
    not input.human_approved
    msg := "HIPAA: PHI access requires explicit human approval"
}

# PHI must not be transmitted without encryption.
deny[msg] {
    lower(input.data_classification) == "phi"
    not input.transmission_encrypted
    msg := "HIPAA: PHI must be transmitted over an encrypted channel"
}

# --- Security Rule ---

# All operations on PHI must produce an audit record.
deny[msg] {
    lower(input.data_classification) == "phi"
    not input.audit_logged
    msg := "HIPAA: all PHI operations must be audit-logged"
}

# Minimum necessary standard — excess data must not be requested.
deny[msg] {
    lower(input.data_classification) == "phi"
    input.excess_data_requested == true
    msg := "HIPAA: minimum necessary standard violated — excess PHI data requested"
}

# Workforce access controls — unauthenticated actors may not touch PHI.
deny[msg] {
    lower(input.data_classification) == "phi"
    not input.actor_authenticated
    msg := "HIPAA: unauthenticated access to PHI is prohibited"
}

# Automatic logoff — sessions idle beyond the threshold must be denied.
deny[msg] {
    input.session_idle_seconds > 900          # 15 minutes
    lower(input.data_classification) == "phi"
    msg := "HIPAA: session idle timeout exceeded — re-authentication required before accessing PHI"
}
