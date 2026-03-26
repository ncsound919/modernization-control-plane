# Module 08 — Data Migration Engine

## Purpose
Avoids the high-risk "rip-and-replace" model with a **tiered migration approach** that separates active data from historical archives. Reduces migration costs by 30–50% while maintaining full compliance and access continuity.

## Tiering Model

| Tier | Data Type | SLO | Storage | Access |
|---|---|---|---|---|
| Tier 1: Active | Recent accounts, patients, transactions, configs | < 100ms | New platform DB (PostgreSQL/Aurora) | Real-time via Sidecar APIs |
| Tier 2: Archive | Historical records, audit logs, closed accounts | < 3s | Object storage (S3/GCS) + query engine | On-demand via Trino/Athena |

## Migration Approach

### Phase 1: Dual-Write
- New writes go to both old and new systems
- Reconciliation job validates consistency
- No data loss risk during transition

### Phase 2: Tier 1 Migration
- Change Data Capture (CDC) streams active data via Debezium
- Kafka as the migration bus (ordered, replayable)
- dbt transforms normalize legacy schemas to target models
- Cutover when new system reaches parity

### Phase 3: Tier 2 Archive
- Batch export historical records to object storage
- Parquet format for efficient querying
- Pre-built query templates for audits, e-discovery, analytics
- Retention and deletion policies per compliance framework

## Tech Stack
- **CDC:** Debezium (MySQL, PostgreSQL, Oracle, SQL Server, DB2)
- **Message Bus:** Apache Kafka
- **Transform:** dbt (data build tool)
- **Query Engine:** Trino / AWS Athena (Tier 2 queries)
- **Archive Storage:** S3 / GCS (Parquet format)
- **Reconciliation:** Custom Go reconciliation service

## Directory Structure
```
08-data-migration/
├── connectors/           # Debezium connector configs per source DB
├── transforms/           # dbt models for schema normalization
├── reconciliation/       # Dual-write consistency checker
├── archive/
│   ├── exporter/         # Batch historical export jobs
│   └── query-templates/  # Pre-built Trino/Athena queries
├── runbooks/             # Per-vertical migration runbooks
│   ├── banking.md
│   └── healthcare.md
└── README.md
```

## Key Metrics
- % of data migrated to Tier 1 vs archived to Tier 2
- Reconciliation error rate (target: < 0.001%)
- Reduction in legacy infra and licensing costs
- Time to first query for Tier 2 archive record

## Status
`Phase 2 — In Development`
