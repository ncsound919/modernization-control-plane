package gdpr

# GDPR (General Data Protection Regulation) enforcement policies.
# Covers the six lawful bases for processing, data subject rights, and
# cross-border transfer restrictions.

default allow = false

allow {
    count(deny) == 0
}

# --- Lawful Basis ---

# Personal data processing requires a valid lawful basis.
deny[msg] {
    is_personal_data
    not has_lawful_basis
    msg := "GDPR: processing personal data requires a valid lawful basis (consent, contract, legal obligation, vital interests, public task, or legitimate interests)"
}

is_personal_data {
    lower(input.data_classification) == "personal"
}
is_personal_data {
    lower(input.data_classification) == "pii"
}

has_lawful_basis {
    input.consent_given == true
}
has_lawful_basis {
    input.lawful_basis != ""
}

# --- Data Minimisation (Art. 5(1)(c)) ---

deny[msg] {
    input.excess_data_requested == true
    msg := "GDPR: data minimisation principle — only data necessary for the stated purpose may be processed"
}

# --- Purpose Limitation (Art. 5(1)(b)) ---

deny[msg] {
    is_personal_data
    input.processing_purpose == ""
    msg := "GDPR: purpose limitation — a specific processing purpose must be declared"
}

# --- Storage Limitation (Art. 5(1)(e)) ---

deny[msg] {
    is_personal_data
    input.retention_days > 365
    not input.extended_retention_justified
    msg := "GDPR: storage limitation — personal data retained beyond 365 days requires justification"
}

# --- Cross-Border Transfer Restriction (Art. 44) ---

deny[msg] {
    is_personal_data
    input.destination_country != ""
    not is_adequate_country(input.destination_country)
    not input.transfer_mechanism_approved
    msg := sprintf("GDPR: cross-border transfer to %v requires an approved transfer mechanism", [input.destination_country])
}

is_adequate_country(country) {
    # EEA member states plus countries with an EU adequacy decision.
    adequate := {"AT","BE","BG","CY","CZ","DE","DK","EE","ES","FI","FR",
                 "GR","HR","HU","IE","IT","LT","LU","LV","MT","NL","PL",
                 "PT","RO","SE","SI","SK","NO","IS","LI",
                 "AD","AR","CA","CH","FO","GB","GG","IL","IM","JE","JP",
                 "NZ","UY"}
    adequate[upper(country)]
}

# --- Right to Erasure (Art. 17) ---

deny[msg] {
    input.action == "retain"
    is_personal_data
    input.erasure_requested == true
    not input.erasure_exception_applies
    msg := "GDPR: right to erasure — data must be deleted upon valid erasure request"
}
