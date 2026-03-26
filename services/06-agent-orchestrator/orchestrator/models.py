from __future__ import annotations

import uuid
from datetime import datetime, timezone
from enum import Enum
from typing import Any

from pydantic import BaseModel, Field


class AgentType(str, Enum):
    discovery = "discovery"
    compliance = "compliance"
    migration = "migration"
    certificate = "certificate"
    cost = "cost"
    test = "test"
    sustainability = "sustainability"


class AgentStatus(str, Enum):
    idle = "idle"
    running = "running"
    completed = "completed"
    failed = "failed"


class WorkflowStatus(str, Enum):
    pending = "pending"
    running = "running"
    completed = "completed"
    failed = "failed"
    cancelled = "cancelled"


class Agent(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    type: AgentType
    status: AgentStatus = AgentStatus.idle
    tenant_id: str = "default"
    slo_latency_ms: int = 30_000
    tool_access: list[str] = Field(default_factory=list)
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    last_active: datetime | None = None


class WorkflowNode(BaseModel):
    agent_type: AgentType
    inputs: dict[str, Any] = Field(default_factory=dict)
    dependencies: list[str] = Field(default_factory=list)


class Workflow(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    name: str
    description: str = ""
    status: WorkflowStatus = WorkflowStatus.pending
    nodes: dict[str, WorkflowNode]
    tenant_id: str = "default"
    created_at: datetime = Field(default_factory=lambda: datetime.now(timezone.utc))
    started_at: datetime | None = None
    completed_at: datetime | None = None
    error: str | None = None


class Job(BaseModel):
    id: str = Field(default_factory=lambda: str(uuid.uuid4()))
    workflow_id: str
    agent_type: AgentType
    node_name: str
    status: AgentStatus = AgentStatus.idle
    inputs: dict[str, Any] = Field(default_factory=dict)
    result: dict[str, Any] | None = None
    error: str | None = None
    started_at: datetime | None = None
    completed_at: datetime | None = None


class WorkflowCreate(BaseModel):
    name: str
    description: str = ""
    nodes: dict[str, WorkflowNode]
    tenant_id: str = "default"


class WorkflowRunResponse(BaseModel):
    workflow_id: str
    status: WorkflowStatus
    message: str
