from __future__ import annotations

from datetime import datetime, timezone
from typing import Any

from fastapi import FastAPI, HTTPException

from orchestrator.models import WorkflowCreate, WorkflowRunResponse, WorkflowStatus
from orchestrator.registry import get_all_agents
from orchestrator.scheduler import scheduler

app = FastAPI(
    title="Agent Orchestrator",
    description="Multi-tenant DAG runtime for coordinating specialized domain agents",
    version="0.1.0",
)


@app.get("/health")
def health() -> dict[str, Any]:
    return {"status": "ok", "service": "agent-orchestrator", "timestamp": datetime.now(timezone.utc).isoformat()}


# ---------------------------------------------------------------------------
# Agents
# ---------------------------------------------------------------------------


@app.get("/api/v1/agents")
def list_agents() -> dict[str, Any]:
    agents = get_all_agents()
    return {"agents": [a.model_dump() for a in agents], "total": len(agents)}


# ---------------------------------------------------------------------------
# Workflows
# ---------------------------------------------------------------------------


@app.get("/api/v1/workflows")
def list_workflows() -> dict[str, Any]:
    workflows = scheduler.list_workflows()
    return {"workflows": [w.model_dump() for w in workflows], "total": len(workflows)}


@app.post("/api/v1/workflows", status_code=201)
def create_workflow(payload: WorkflowCreate) -> dict[str, Any]:
    wf = scheduler.create_workflow(payload)
    return wf.model_dump()


@app.get("/api/v1/workflows/{workflow_id}")
def get_workflow(workflow_id: str) -> dict[str, Any]:
    wf = scheduler.get_workflow(workflow_id)
    if wf is None:
        raise HTTPException(status_code=404, detail=f"Workflow '{workflow_id}' not found")
    jobs = scheduler.jobs_for_workflow(workflow_id)
    data = wf.model_dump()
    data["jobs"] = [j.model_dump() for j in jobs]
    return data


@app.post("/api/v1/workflows/{workflow_id}/run")
async def run_workflow(workflow_id: str) -> WorkflowRunResponse:
    wf = scheduler.get_workflow(workflow_id)
    if wf is None:
        raise HTTPException(status_code=404, detail=f"Workflow '{workflow_id}' not found")
    wf = await scheduler.run_workflow(workflow_id)
    return WorkflowRunResponse(
        workflow_id=workflow_id,
        status=wf.status,
        message="Workflow execution started" if wf.status == WorkflowStatus.running else wf.status.value,
    )


@app.post("/api/v1/workflows/{workflow_id}/cancel")
def cancel_workflow(workflow_id: str) -> dict[str, Any]:
    wf = scheduler.cancel_workflow(workflow_id)
    if wf is None:
        raise HTTPException(status_code=404, detail=f"Workflow '{workflow_id}' not found")
    return {"workflow_id": workflow_id, "status": wf.status, "message": "Workflow cancelled"}


# ---------------------------------------------------------------------------
# Jobs
# ---------------------------------------------------------------------------


@app.get("/api/v1/jobs")
def list_jobs() -> dict[str, Any]:
    jobs = scheduler.list_jobs()
    return {"jobs": [j.model_dump() for j in jobs], "total": len(jobs)}


@app.get("/api/v1/jobs/{job_id}")
def get_job(job_id: str) -> dict[str, Any]:
    job = scheduler.get_job(job_id)
    if job is None:
        raise HTTPException(status_code=404, detail=f"Job '{job_id}' not found")
    return job.model_dump()


# ---------------------------------------------------------------------------
# Metrics
# ---------------------------------------------------------------------------


@app.get("/api/v1/metrics")
def get_metrics() -> dict[str, Any]:
    return scheduler.metrics()
