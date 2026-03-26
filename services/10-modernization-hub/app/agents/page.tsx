import StatusBadge from "@/components/StatusBadge";

interface Agent {
  id: string;
  name: string;
  type: string;
  description: string;
  status: "active" | "idle";
  runsToday: number;
}

interface Workflow {
  id: string;
  name: string;
  agents: string[];
  status: "active" | "idle";
  progress: number;
  startedAt: string;
  target: string;
}

const agents: Agent[] = [
  {
    id: "discovery-agent",
    name: "Discovery Agent",
    type: "Module 01",
    description: "Scans infrastructure and registers services, certs, and dependencies",
    status: "active",
    runsToday: 48,
  },
  {
    id: "clm-agent",
    name: "CLM Agent",
    type: "Module 03",
    description: "Automates certificate lifecycle: issuance, renewal, and revocation",
    status: "active",
    runsToday: 22,
  },
  {
    id: "migration-agent",
    name: "Migration Agent",
    type: "Module 05",
    description: "Generates and executes migration runbooks for legacy services",
    status: "active",
    runsToday: 7,
  },
  {
    id: "governance-agent",
    name: "Governance Agent",
    type: "Module 07",
    description: "Evaluates policy compliance and triggers kill switches on violation",
    status: "active",
    runsToday: 31,
  },
  {
    id: "debt-scorer",
    name: "Debt Scorer",
    type: "Module 01",
    description: "Computes technical debt risk scores from graph analysis",
    status: "idle",
    runsToday: 4,
  },
  {
    id: "sustainability-agent",
    name: "Sustainability Agent",
    type: "Module 08",
    description: "Tracks gCO2e emissions per workflow and generates ESG reports",
    status: "idle",
    runsToday: 2,
  },
  {
    id: "pricing-agent",
    name: "Pricing Agent",
    type: "Module 09",
    description: "Calculates outcome-based billing metrics and ROI deltas",
    status: "idle",
    runsToday: 1,
  },
];

const workflows: Workflow[] = [
  {
    id: "wf-001",
    name: "Full Legacy Auth Migration",
    agents: ["Discovery Agent", "Migration Agent", "Governance Agent"],
    status: "active",
    progress: 68,
    startedAt: "09:15 today",
    target: "legacy-auth-service",
  },
  {
    id: "wf-002",
    name: "Bulk Cert Renewal — Finance",
    agents: ["CLM Agent", "Governance Agent"],
    status: "active",
    progress: 91,
    startedAt: "11:02 today",
    target: "billing.api.internal",
  },
  {
    id: "wf-003",
    name: "Billing Monolith Debt Reduction",
    agents: ["Debt Scorer", "Migration Agent"],
    status: "active",
    progress: 22,
    startedAt: "13:45 today",
    target: "billing-monolith",
  },
  {
    id: "wf-004",
    name: "Sustainability Audit — Q3",
    agents: ["Sustainability Agent", "Governance Agent"],
    status: "active",
    progress: 55,
    startedAt: "14:00 today",
    target: "all services",
  },
];

export default function AgentsPage() {
  const activeAgents = agents.filter((a) => a.status === "active").length;
  const totalRuns = agents.reduce((sum, a) => sum + a.runsToday, 0);

  return (
    <div>
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-slate-800">
          Agent Workflow Composer
        </h2>
        <p className="text-slate-500 mt-1">
          Available agents and active multi-agent workflows
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-3 gap-4 mb-8">
        {[
          { label: "Agent Types", value: agents.length, color: "text-indigo-700" },
          { label: "Active Agents", value: activeAgents, color: "text-emerald-700" },
          { label: "Runs Today", value: totalRuns, color: "text-blue-700" },
        ].map((s) => (
          <div key={s.label} className="bg-white border border-slate-200 rounded-xl p-5 shadow-sm">
            <p className={`text-3xl font-bold ${s.color}`}>{s.value}</p>
            <p className="text-xs text-slate-500 mt-1">{s.label}</p>
          </div>
        ))}
      </div>

      {/* Active Workflows */}
      <h3 className="font-semibold text-slate-700 mb-4">Active Workflows</h3>
      <div className="grid grid-cols-1 gap-4 mb-8 xl:grid-cols-2">
        {workflows.map((wf) => (
          <div key={wf.id} className="bg-white border border-slate-200 rounded-xl shadow-sm p-5">
            <div className="flex items-start justify-between mb-3">
              <div>
                <p className="font-semibold text-slate-800">{wf.name}</p>
                <p className="text-xs text-slate-500 mt-0.5">
                  Target: <span className="font-mono text-indigo-600">{wf.target}</span> · Started {wf.startedAt}
                </p>
              </div>
              <StatusBadge status={wf.status} />
            </div>

            {/* Progress */}
            <div className="mb-3">
              <div className="flex justify-between text-xs text-slate-500 mb-1">
                <span>Progress</span>
                <span>{wf.progress}%</span>
              </div>
              <div className="h-2 bg-slate-100 rounded-full overflow-hidden">
                <div
                  className="h-full bg-indigo-500 rounded-full transition-all"
                  style={{ width: `${wf.progress}%` }}
                />
              </div>
            </div>

            {/* Agent chain */}
            <div className="flex items-center gap-2 flex-wrap">
              {wf.agents.map((agent, i) => (
                <span key={agent} className="flex items-center gap-1">
                  <span className="bg-indigo-50 text-indigo-700 border border-indigo-200 text-xs rounded px-2 py-0.5 font-medium">
                    {agent}
                  </span>
                  {i < wf.agents.length - 1 && (
                    <span className="text-slate-300">→</span>
                  )}
                </span>
              ))}
            </div>
          </div>
        ))}
      </div>

      {/* Agent catalog */}
      <h3 className="font-semibold text-slate-700 mb-4">Agent Catalog</h3>
      <div className="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              {["Agent", "Module", "Description", "Status", "Runs Today"].map((h) => (
                <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wide">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {agents.map((agent) => (
              <tr key={agent.id} className="hover:bg-slate-50 transition-colors">
                <td className="px-4 py-3 font-medium text-slate-800">{agent.name}</td>
                <td className="px-4 py-3">
                  <span className="bg-indigo-100 text-indigo-700 text-xs rounded px-2 py-0.5 font-semibold">
                    {agent.type}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-500 text-xs max-w-xs">{agent.description}</td>
                <td className="px-4 py-3">
                  <StatusBadge status={agent.status} />
                </td>
                <td className="px-4 py-3 text-slate-700 font-semibold">{agent.runsToday}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
