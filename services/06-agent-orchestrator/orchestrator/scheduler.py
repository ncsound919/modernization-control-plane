"""In-memory DAG scheduler.

Workflows are stored in a dict keyed by workflow id.  When a workflow is
triggered, the scheduler resolves execution order via topological sort and
runs each node sequentially (simulating realistic agent execution).  Because
this is a simulation layer, actual I/O against Redis / Kafka / Postgres is
intentionally omitted — the service can start standalone without external
dependencies.
"""
from __future__ import annotations

import asyncio
import random
import uuid
from collections import deque
from datetime import datetime, timezone
from typing import Any

from orchestrator.models import (
    AgentStatus,
    AgentType,
    Job,
    Workflow,
    WorkflowCreate,
    WorkflowNode,
    WorkflowStatus,
)

# ---------------------------------------------------------------------------
# Pre-built workflow templates
# ---------------------------------------------------------------------------

WORKFLOW_TEMPLATES: dict[str, dict[str, Any]] = {
    "full-modernization": {
        "name": "Full Modernization",
        "description": "End-to-end modernization pipeline: discovery → compliance → migration → test → cost → sustainability",
        "nodes": {
            "discover": WorkflowNode(agent_type=AgentType.discovery, inputs={"scan_depth": "full"}),
            "compliance_check": WorkflowNode(
                agent_type=AgentType.compliance,
                inputs={"frameworks": ["HIPAA", "GDPR", "SOC2"]},
                dependencies=["discover"],
            ),
            "migrate": WorkflowNode(
                agent_type=AgentType.migration,
                inputs={"strategy": "blue-green"},
                dependencies=["compliance_check"],
            ),
            "test_suite": WorkflowNode(
                agent_type=AgentType.test,
                inputs={"coverage_threshold": 80},
                dependencies=["migrate"],
            ),
            "cost_analysis": WorkflowNode(
                agent_type=AgentType.cost,
                inputs={"forecast_months": 12},
                dependencies=["migrate"],
            ),
            "sustainability": WorkflowNode(
                agent_type=AgentType.sustainability,
                inputs={"baseline": "current"},
                dependencies=["migrate"],
            ),
        },
    },
    "cert-rotation": {
        "name": "Certificate Rotation",
        "description": "Automated certificate discovery, validation, and rotation",
        "nodes": {
            "discover_certs": WorkflowNode(
                agent_type=AgentType.discovery,
                inputs={"target": "certificates"},
            ),
            "rotate": WorkflowNode(
                agent_type=AgentType.certificate,
                inputs={"provider": "vault"},
                dependencies=["discover_certs"],
            ),
            "compliance_verify": WorkflowNode(
                agent_type=AgentType.compliance,
                inputs={"frameworks": ["SOC2"]},
                dependencies=["rotate"],
            ),
        },
    },
    "compliance-check": {
        "name": "Compliance Check",
        "description": "Standalone compliance audit across all supported frameworks",
        "nodes": {
            "scan": WorkflowNode(
                agent_type=AgentType.discovery,
                inputs={"scan_depth": "shallow"},
            ),
            "audit": WorkflowNode(
                agent_type=AgentType.compliance,
                inputs={"frameworks": ["HIPAA", "GDPR", "SOC2"]},
                dependencies=["scan"],
            ),
            "cost_impact": WorkflowNode(
                agent_type=AgentType.cost,
                inputs={"scope": "compliance"},
                dependencies=["audit"],
            ),
        },
    },
}

# ---------------------------------------------------------------------------
# Mock result generators (per agent type)
# ---------------------------------------------------------------------------

_MOCK_RESULTS: dict[AgentType, dict[str, Any]] = {
    AgentType.discovery: {
        "services_scanned": 42,
        "tech_debt_score": 67.3,
        "critical_issues": 5,
        "recommendations": ["upgrade-java-17", "remove-deprecated-apis", "add-observability"],
    },
    AgentType.compliance: {
        "frameworks_checked": ["HIPAA", "GDPR", "SOC2"],
        "passed": 18,
        "failed": 2,
        "warnings": 4,
        "report_url": "https://compliance.internal/report/mock-001",
    },
    AgentType.migration: {
        "tables_migrated": 15,
        "rows_transferred": 1_234_567,
        "duration_s": 47,
        "rollback_available": True,
    },
    AgentType.certificate: {
        "certs_found": 12,
        "certs_rotated": 12,
        "expiring_soon": 0,
        "next_rotation": "2025-01-01T00:00:00Z",
    },
    AgentType.cost: {
        "current_monthly_usd": 18_420,
        "projected_monthly_usd": 14_300,
        "savings_pct": 22.4,
        "top_cost_drivers": ["compute", "data-transfer", "storage"],
    },
    AgentType.test: {
        "tests_run": 2_847,
        "passed": 2_831,
        "failed": 16,
        "coverage_pct": 84.2,
        "regression_detected": False,
    },
    AgentType.sustainability: {
        "baseline_kg_co2e_month": 340,
        "projected_kg_co2e_month": 251,
        "reduction_pct": 26.2,
        "green_score": "B+",
    },
}

# Simulated execution time range per agent (seconds)
_EXEC_DELAY: dict[AgentType, tuple[float, float]] = {
    AgentType.discovery: (0.3, 0.8),
    AgentType.compliance: (0.2, 0.5),
    AgentType.migration: (0.5, 1.2),
    AgentType.certificate: (0.1, 0.3),
    AgentType.cost: (0.2, 0.4),
    AgentType.test: (0.4, 1.0),
    AgentType.sustainability: (0.2, 0.5),
}


# ---------------------------------------------------------------------------
# DAG Scheduler
# ---------------------------------------------------------------------------


class DAGScheduler:
    def __init__(self) -> None:
        self._workflows: dict[str, Workflow] = {}
        self._jobs: dict[str, Job] = {}
        # Seed pre-built templates so they appear in the workflows list
        for template_id, template in WORKFLOW_TEMPLATES.items():
            wf = Workflow(
                id=template_id,
                name=template["name"],
                description=template["description"],
                nodes=template["nodes"],
            )
            self._workflows[wf.id] = wf

    # ------------------------------------------------------------------
    # Workflow CRUD
    # ------------------------------------------------------------------

    def create_workflow(self, payload: WorkflowCreate) -> Workflow:
        wf = Workflow(
            name=payload.name,
            description=payload.description,
            nodes=payload.nodes,
            tenant_id=payload.tenant_id,
        )
        self._workflows[wf.id] = wf
        return wf

    def get_workflow(self, workflow_id: str) -> Workflow | None:
        return self._workflows.get(workflow_id)

    def list_workflows(self) -> list[Workflow]:
        return list(self._workflows.values())

    def cancel_workflow(self, workflow_id: str) -> Workflow | None:
        wf = self._workflows.get(workflow_id)
        if wf is None:
            return None
        if wf.status in (WorkflowStatus.pending, WorkflowStatus.running):
            wf.status = WorkflowStatus.cancelled
            wf.completed_at = datetime.now(timezone.utc)
        return wf

    # ------------------------------------------------------------------
    # Job helpers
    # ------------------------------------------------------------------

    def list_jobs(self) -> list[Job]:
        return list(self._jobs.values())

    def get_job(self, job_id: str) -> Job | None:
        return self._jobs.get(job_id)

    def jobs_for_workflow(self, workflow_id: str) -> list[Job]:
        return [j for j in self._jobs.values() if j.workflow_id == workflow_id]

    # ------------------------------------------------------------------
    # Execution
    # ------------------------------------------------------------------

    async def run_workflow(self, workflow_id: str) -> Workflow | None:
        wf = self._workflows.get(workflow_id)
        if wf is None:
            return None
        if wf.status == WorkflowStatus.running:
            return wf
        if wf.status == WorkflowStatus.cancelled:
            return wf

        wf.status = WorkflowStatus.running
        wf.started_at = datetime.now(timezone.utc)

        # Fire-and-forget: run the DAG in the background
        asyncio.create_task(self._execute_dag(wf))
        return wf

    async def _execute_dag(self, wf: Workflow) -> None:
        try:
            order = _topological_sort(wf.nodes)
        except ValueError as exc:
            wf.status = WorkflowStatus.failed
            wf.error = str(exc)
            wf.completed_at = datetime.now(timezone.utc)
            return

        completed_nodes: set[str] = set()
        node_results: dict[str, dict[str, Any]] = {}

        for node_name in order:
            # Check if workflow was cancelled mid-run
            if wf.status == WorkflowStatus.cancelled:
                return

            node: WorkflowNode = wf.nodes[node_name]
            job = Job(
                id=str(uuid.uuid4()),
                workflow_id=wf.id,
                agent_type=node.agent_type,
                node_name=node_name,
                inputs=node.inputs,
                status=AgentStatus.running,
                started_at=datetime.now(timezone.utc),
            )
            self._jobs[job.id] = job

            low, high = _EXEC_DELAY[node.agent_type]
            await asyncio.sleep(random.uniform(low, high))

            job.result = dict(_MOCK_RESULTS[node.agent_type])
            job.result["node_name"] = node_name
            job.result["upstream"] = {dep: node_results.get(dep) for dep in node.dependencies}
            job.status = AgentStatus.completed
            job.completed_at = datetime.now(timezone.utc)

            completed_nodes.add(node_name)
            node_results[node_name] = job.result

        wf.status = WorkflowStatus.completed
        wf.completed_at = datetime.now(timezone.utc)

    # ------------------------------------------------------------------
    # Metrics
    # ------------------------------------------------------------------

    def metrics(self) -> dict[str, Any]:
        workflows = list(self._workflows.values())
        jobs = list(self._jobs.values())
        return {
            "workflows": {
                "total": len(workflows),
                "by_status": _count_by(workflows, "status"),
            },
            "jobs": {
                "total": len(jobs),
                "by_status": _count_by(jobs, "status"),
                "by_agent_type": _count_by(jobs, "agent_type"),
            },
            "agents_registered": 7,
        }


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------


def _topological_sort(nodes: dict[str, WorkflowNode]) -> list[str]:
    """Return node names in a valid topological execution order."""
    in_degree: dict[str, int] = {n: 0 for n in nodes}
    adjacency: dict[str, list[str]] = {n: [] for n in nodes}

    for name, node in nodes.items():
        for dep in node.dependencies:
            if dep not in nodes:
                raise ValueError(f"Node '{name}' depends on unknown node '{dep}'")
            adjacency[dep].append(name)
            in_degree[name] += 1

    queue: deque[str] = deque(n for n, deg in in_degree.items() if deg == 0)
    order: list[str] = []

    while queue:
        current = queue.popleft()
        order.append(current)
        for neighbor in adjacency[current]:
            in_degree[neighbor] -= 1
            if in_degree[neighbor] == 0:
                queue.append(neighbor)

    if len(order) != len(nodes):
        raise ValueError("Workflow graph contains a cycle")
    return order


def _count_by(items: list[Any], field: str) -> dict[str, int]:
    counts: dict[str, int] = {}
    for item in items:
        raw = getattr(item, field, "unknown")
        key = raw.value if hasattr(raw, "value") else str(raw)
        counts[key] = counts.get(key, 0) + 1
    return counts


# Singleton instance shared across the application
scheduler = DAGScheduler()
