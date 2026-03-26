# Module 06 — Agent Orchestrator

## Purpose
"Kubernetes for AI agents" — a multi-tenant runtime that coordinates specialized domain agents working together on complex modernization goals. Shifts the platform from simple chatbots to **Multi-Agent Systems (MAS)** with defined SLOs, tool access policies, and execution guarantees.

## Core Concepts

### Agent Types
| Agent | Role |
|---|---|
| Discovery Agent | Scans and scores technical debt (feeds Module 01) |
| Compliance Agent | Validates changes against HIPAA, GDPR, SOC 2 |
| Migration Agent | Plans and executes tiered data migrations |
| Certificate Agent | Monitors and triggers CLM rotations |
| Cost Agent | Tracks and forecasts modernization costs |
| Test Agent | Generates and runs regression tests |
| Sustainability Agent | Calculates emissions delta for changes |

### Orchestration Primitives
- **Agent Spec:** Defines model, tools, data access policy, latency SLO
- **Agent Job:** A workflow instance with start condition and completion criteria
- **DAG Scheduler:** Routes jobs to agents based on skill, data locality, compliance constraints
- **State Machine:** Workflows run as explicit state machines, not free-form prompts

## Tech Stack
- **Orchestration:** Python 3.11+, LangGraph (state machine workflows)
- **Agent Framework:** CrewAI for multi-agent coordination
- **State Store:** Redis (in-flight state), PostgreSQL (job history)
- **Message Bus:** Apache Kafka (agent-to-agent events)
- **API:** REST + gRPC agent control plane
- **LLM Integration:** OpenAI, Anthropic, local Ollama (configurable per agent)

## Directory Structure
```
06-agent-orchestrator/
├── orchestrator/
│   ├── scheduler/        # DAG scheduler and job router
│   ├── state/            # Redis/PG state management
│   └── api/              # Control plane REST/gRPC API
├── agents/
│   ├── discovery/
│   ├── compliance/
│   ├── migration/
│   ├── certificate/
│   ├── cost/
│   ├── test/
│   └── sustainability/
├── workflows/            # Pre-built modernization workflow templates
├── tools/                # Shared tools (DB connectors, API clients)
├── Dockerfile
├── pyproject.toml
└── README.md
```

## Key Metrics
- Mean time from discovery to executed change
- % of modernization changes driven or assisted by agent workflows
- Agent SLO adherence rate (latency, cost, accuracy)

## Status
`Phase 4 — Planned`
