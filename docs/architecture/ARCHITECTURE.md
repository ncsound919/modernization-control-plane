# System Architecture

## Overview

The Modernization Control Plane is a **polyglot microservices platform** with a modular service architecture. Each service is independently deployable and communicates via REST/gRPC APIs and Kafka event streams.

## Service Communication

```
                    +---------------------------+
                    |   Modernization Hub (10)  |  <-- Next.js UI
                    +---------------------------+
                               |
                           tRPC / REST
                               |
           +-------------------+-------------------+
           |                   |                   |
   +-------+-------+   +-------+-------+   +-------+-------+
   | Discovery (01)|   | CLM Svc  (03) |   | Governance(07)|
   +---------------+   +---------------+   +---------------+
           |                   |                   |
           |           +-------+-------+           |
           |           |Agent Orch (06)|           |
           |           +-------+-------+           |
           |                   |                   |
   +-------+-------+   +-------+-------+   +-------+-------+
   |Sidecar GW (02)|   |Data Migr (08) |   |Sustain   (09) |
   +---------------+   +---------------+   +---------------+
           |                   |
   +-------+-------+   +-------+-------+
   |Browser Cnt(04)|   |Proj Import(05)|
   +---------------+   +---------------+
```

## Data Flow: Discovery to Modernization

1. **Discovery Engine** scans cloud infrastructure → writes to Discovery Graph (Neo4j)
2. **CLM Service** reads cert inventory from graph → auto-rotates certs via ACME
3. **Sidecar Gateway** wraps legacy systems → exposes stable domain APIs
4. **Agent Orchestrator** reads graph risk scores → dispatches modernization workflows
5. **Governance Engine** intercepts all agent actions → policy check → audit log
6. **Data Migration** moves data through Tier 1/2 → feeds Sidecar Gateway
7. **Sustainability** measures emissions of all above → reports to Hub
8. **Modernization Hub** visualizes everything → outcome pricing hooks

## Deployment Architecture

### Production (Kubernetes)
- Each service runs as a K8s Deployment with HPA
- Services communicate via cluster-internal DNS
- Sidecar Gateway deployed close to legacy systems (on-prem or VPC)
- Discovery agents deployed as DaemonSets or VM agents

### Local Development (Docker Compose)
- All services available via `docker compose up -d`
- Shared infrastructure: PostgreSQL, Redis, Kafka, Neo4j
- Hot-reload for Go (Air) and TypeScript (Next.js dev server)

## Shared Infrastructure

| Service | Purpose |
|---|---|
| PostgreSQL | Primary relational DB (CLM inventory, governance audit, agent jobs) |
| Neo4j | Discovery Graph (debt heatmap, asset relationships) |
| Redis | Agent in-flight state, rate limiting, caching |
| Apache Kafka | Event bus (agent events, CDC streams, governance events) |
| HashiCorp Vault | Secret management (private keys, API credentials) |
| OpenTelemetry Collector | Unified trace, metric, and log pipeline |

## Security

- mTLS between all services
- JWT-based API auth (OIDC provider for enterprise SSO)
- All data at rest encrypted (AES-256)
- Private keys never leave Vault
- OPA policy checks at service boundaries

## ADRs (Architecture Decision Records)

See `docs/architecture/adr/` for individual decision records.

| ADR | Decision |
|---|---|
| ADR-001 | Go for backend services (performance, deployment simplicity) |
| ADR-002 | Python for agent orchestrator (LangGraph ecosystem) |
| ADR-003 | Next.js 15 for Hub (React ecosystem, SSR for enterprise auth) |
| ADR-004 | OPA for policy engine (standard, auditable, versionable policies) |
| ADR-005 | Kafka for event bus (ordered, replayable, battle-tested at scale) |
| ADR-006 | Neo4j for discovery graph (native graph traversal for debt mapping) |
