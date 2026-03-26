from __future__ import annotations

from orchestrator.models import Agent, AgentType

_TOOL_ACCESS: dict[AgentType, list[str]] = {
    AgentType.discovery: ["repo_scanner", "dependency_graph", "sonar_api"],
    AgentType.compliance: ["policy_db", "audit_log", "hipaa_ruleset", "gdpr_ruleset", "soc2_ruleset"],
    AgentType.migration: ["db_connector", "schema_differ", "etl_runner", "rollback_engine"],
    AgentType.certificate: ["cert_store", "acme_api", "vault_api", "rotation_scheduler"],
    AgentType.cost: ["cloud_billing_api", "resource_inventory", "cost_forecaster"],
    AgentType.test: ["test_runner", "coverage_reporter", "regression_suite"],
    AgentType.sustainability: ["emissions_calculator", "resource_monitor", "carbon_api"],
}

_SLO_LATENCY_MS: dict[AgentType, int] = {
    AgentType.discovery: 60_000,
    AgentType.compliance: 15_000,
    AgentType.migration: 120_000,
    AgentType.certificate: 10_000,
    AgentType.cost: 20_000,
    AgentType.test: 90_000,
    AgentType.sustainability: 25_000,
}

# One registered agent per type (singleton pool for simulation)
_REGISTRY: dict[AgentType, Agent] = {
    agent_type: Agent(
        type=agent_type,
        tool_access=_TOOL_ACCESS[agent_type],
        slo_latency_ms=_SLO_LATENCY_MS[agent_type],
    )
    for agent_type in AgentType
}


def get_all_agents() -> list[Agent]:
    return list(_REGISTRY.values())


def get_agent(agent_type: AgentType) -> Agent | None:
    return _REGISTRY.get(agent_type)
