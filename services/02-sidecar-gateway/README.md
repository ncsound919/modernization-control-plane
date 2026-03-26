# Module 02 — Sidecar Gateway

## Purpose
Acts as a **translation bridge** between modern cloud applications and legacy record-keeping systems. Enables new features to ship in weeks rather than the 18 months typical of legacy core releases.

## Architecture

### Layer 1 — Stable API Contracts
Opinionated REST/GraphQL/gRPC interfaces aligned to domain language (Accounts, Claims, Patients, Orders). New apps talk only to this layer — contracts are versioned and immutable.

### Layer 2 — Legacy Protocol Adapters
Pluggable adapters for each legacy backend:
- COBOL batch via JCL/MQ
- IBM MQ / RabbitMQ message queues
- SFTP flat-file exchange
- HL7 v2 / FHIR (healthcare)
- Direct DB2 / VSAM query adapters

## Tech Stack
- **Gateway:** Envoy Proxy / Kong
- **Adapter SDK:** Go 1.22+
- **Contracts:** Protobuf + OpenAPI 3.1
- **Auth:** JWT + mTLS between gateway and adapters
- **Observability:** OpenTelemetry traces and metrics

## Directory Structure
```
02-sidecar-gateway/
├── gateway/              # Envoy/Kong config and filters
├── adapters/
│   ├── cobol/            # COBOL/MQ adapter
│   ├── hl7/              # HL7 v2 / FHIR adapter
│   ├── sftp/             # Flat-file SFTP adapter
│   └── db2/              # DB2 / VSAM adapter
├── contracts/            # OpenAPI + Protobuf definitions
├── middleware/           # Auth, rate limiting, circuit breakers
├── Dockerfile
├── go.mod
└── README.md
```

## Key Metrics
- Median time to expose a new legacy function via the sidecar
- Number of modern use cases shipped without touching the core
- p99 latency overhead introduced by the gateway layer

## Status
`Phase 1-2 — In Development`
