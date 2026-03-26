# Module 01 — Discovery Engine

## Purpose
Builds a **Technical Debt Graph** by scanning multi-cloud environments (AWS, Azure, GCP) and on-prem infrastructure. Outputs a risk-scored heatmap that drives the rest of the modernization workflow.

## Capabilities
- TLS/SSL certificate discovery across load balancers, gateways, internal services, and edge endpoints
- Legacy system fingerprinting: COBOL batch jobs, mainframe queues, VSAM/DB2, EMR/EHR vendor signatures, HL7 v2, FHIR endpoints
- Cloud asset inventory ingestion via provider APIs
- Centralized Discovery Graph (Nodes: apps, DBs, certs, queues; Edges: calls, ownership, risk)

## Tech Stack
- **Language:** Go 1.22+
- **Graph DB:** Neo4j / AWS Neptune
- **Cloud SDKs:** AWS SDK v2, Azure SDK for Go, Google Cloud Go SDK
- **Cert Scanning:** Custom TLS probe + Nmap scripting engine integration
- **Deployment:** Kubernetes DaemonSet (cloud), Windows/Linux agent service (on-prem)

## Directory Structure
```
01-discovery-engine/
├── cmd/
│   └── scanner/          # Main scanner entrypoint
├── internal/
│   ├── cloud/            # AWS, Azure, GCP connectors
│   ├── legacy/           # COBOL, mainframe, EMR/EHR fingerprinters
│   ├── certs/            # TLS/SSL certificate scanner
│   ├── graph/            # Discovery Graph builder and writer
│   └── scoring/          # Risk scoring engine
├── config/               # Scan profiles and target definitions
├── Dockerfile
├── go.mod
└── README.md
```

## Key Metrics
- % of estate discovered
- Certs and legacy systems mapped vs. estimated total
- Time to first debt heatmap for a new customer

## Status
`Phase 1 — In Development`
