export const config = {
  discoveryApi: process.env.DISCOVERY_API ?? "http://discovery-engine:8080",
  clmApi: process.env.CLM_API ?? "http://clm-service:8080",
  governanceApi: process.env.GOVERNANCE_API ?? "http://governance-engine:8080",
  agentApi: process.env.AGENT_API ?? "http://agent-orchestrator:8080",
};
