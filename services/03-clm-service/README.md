# Module 03 — Certificate Lifecycle Manager (CLM)

## Purpose
Automates TLS/SSL certificate issuance, renewal, and revocation using the **ACME protocol**. Eliminates manual certificate management as CA/B Forum validity windows shrink from 398 → 200 → 100 → 47 days between 2026–2029.

## Why This Is Critical

| Year | Max Cert Validity | Renewals/Year (per cert) |
|---|---|---|
| 2025 | 398 days | ~1 |
| 2026 (Mar 15) | 200 days | ~2 |
| 2027 | 100 days | ~4 |
| 2029 | 47 days | ~8 |

Manual spreadsheet tracking fails at this renewal frequency. CLM automation is operationally mandatory.

## Capabilities
- ACME v2 client for Let's Encrypt and enterprise CAs (DigiCert, Sectigo, internal EJBCA)
- DNS-01 and HTTP-01 challenge resolvers
- Certificate inventory synced from Module 01 (Discovery Engine)
- Policy engine: auto-rotate N days before expiry, per segment
- Deployment integrations: Kubernetes Secrets, AWS ACM, Azure Key Vault, Nginx/Envoy reload
- Alert and block CI/CD deploys when certs are near expiry

## Tech Stack
- **Language:** Go 1.22+
- **ACME library:** `golang.org/x/crypto/acme`
- **Storage:** PostgreSQL (cert inventory), Vault (private keys)
- **Integrations:** cert-manager (K8s), AWS ACM, Azure Key Vault

## Directory Structure
```
03-clm-service/
├── cmd/
│   └── clm/              # Main CLM service entrypoint
├── internal/
│   ├── acme/             # ACME protocol client
│   ├── challenges/       # DNS-01, HTTP-01 resolvers
│   ├── inventory/        # Cert inventory DB layer
│   ├── policy/           # Rotation policy engine
│   └── deploy/           # K8s, ACM, Key Vault integrations
├── Dockerfile
├── go.mod
└── README.md
```

## Key Metrics
- % of discovered certs under automated renewal
- Cert-expiry-related outages prevented
- Time from cert expiry alert to renewed deployment

## Status
`Phase 1 — Priority Build (March 2026 deadline)`
