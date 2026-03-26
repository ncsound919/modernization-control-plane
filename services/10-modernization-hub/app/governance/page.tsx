import StatusBadge from "@/components/StatusBadge";

interface Policy {
  name: string;
  framework: string;
  scope: string;
  status: "compliant" | "violation";
  lastEvaluated: string;
  findings: number;
}

interface KillSwitch {
  id: string;
  name: string;
  scope: string;
  enabled: boolean;
  lastTriggered: string;
}

interface AuditEntry {
  time: string;
  actor: string;
  action: string;
  resource: string;
  outcome: "success" | "blocked" | "warning";
}

const policies: Policy[] = [
  { name: "Data Encryption at Rest", framework: "HIPAA", scope: "All services", status: "compliant", lastEvaluated: "5 min ago", findings: 0 },
  { name: "PHI Access Controls", framework: "HIPAA", scope: "Healthcare domain", status: "compliant", lastEvaluated: "5 min ago", findings: 0 },
  { name: "Right to Erasure (Art. 17)", framework: "GDPR", scope: "Customer data", status: "compliant", lastEvaluated: "12 min ago", findings: 0 },
  { name: "Data Portability (Art. 20)", framework: "GDPR", scope: "User profile API", status: "compliant", lastEvaluated: "12 min ago", findings: 0 },
  { name: "Consent Management", framework: "GDPR", scope: "Frontend services", status: "violation", lastEvaluated: "1 hr ago", findings: 2 },
  { name: "Logical Access Controls", framework: "SOC 2", scope: "All services", status: "compliant", lastEvaluated: "30 min ago", findings: 0 },
  { name: "Change Management", framework: "SOC 2", scope: "CI/CD pipeline", status: "compliant", lastEvaluated: "30 min ago", findings: 0 },
  { name: "Availability SLO ≥ 99.9%", framework: "SOC 2", scope: "Tier 1 services", status: "compliant", lastEvaluated: "1 min ago", findings: 0 },
];

const killSwitches: KillSwitch[] = [
  { id: "ks-001", name: "Stop all migrations", scope: "Global", enabled: false, lastTriggered: "Never" },
  { id: "ks-002", name: "Halt CLM auto-renewal", scope: "Finance domain", enabled: false, lastTriggered: "3 days ago" },
  { id: "ks-003", name: "Pause acme-corp workflows", scope: "Tenant: acme-corp", enabled: true, lastTriggered: "15 min ago" },
  { id: "ks-004", name: "Block external cert CA", scope: "Global", enabled: false, lastTriggered: "Never" },
];

const auditLog: AuditEntry[] = [
  { time: "14:32", actor: "governance-agent", action: "POLICY_EVALUATED", resource: "consent-management", outcome: "warning" },
  { time: "14:28", actor: "admin@platform.io", action: "KILL_SWITCH_ACTIVATED", resource: "ks-003 / acme-corp", outcome: "success" },
  { time: "14:15", actor: "clm-agent", action: "CERT_RENEWED", resource: "api.payments.internal", outcome: "success" },
  { time: "14:02", actor: "migration-agent", action: "RUNBOOK_EXECUTED", resource: "legacy-auth-service step 3/5", outcome: "success" },
  { time: "13:55", actor: "governance-agent", action: "POLICY_EVALUATED", resource: "GDPR / user-profile-api", outcome: "success" },
  { time: "13:40", actor: "admin@platform.io", action: "POLICY_UPDATED", resource: "data-encryption-at-rest", outcome: "success" },
  { time: "13:22", actor: "discovery-agent", action: "ASSET_REGISTERED", resource: "svc-inventory-v2", outcome: "success" },
  { time: "13:00", actor: "pricing-agent", action: "OUTCOME_RECORDED", resource: "acme-corp billing cycle", outcome: "success" },
  { time: "12:45", actor: "governance-agent", action: "VIOLATION_DETECTED", resource: "consent-management / devportal", outcome: "blocked" },
  { time: "12:30", actor: "clm-agent", action: "CERT_EXPIRED_ALERT", resource: "staging-api.internal", outcome: "warning" },
];

const outcomeColors: Record<string, string> = {
  success: "text-emerald-700 bg-emerald-50",
  blocked: "text-red-700 bg-red-50",
  warning: "text-amber-700 bg-amber-50",
};

export default function GovernancePage() {
  const violations = policies.filter((p) => p.status === "violation").length;
  const activeKS = killSwitches.filter((k) => k.enabled).length;

  return (
    <div>
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-slate-800">Governance Console</h2>
        <p className="text-slate-500 mt-1">
          Policy posture, kill switches, and audit log across all tenants
        </p>
      </div>

      {/* Stats */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        {[
          { label: "Policies Active", value: policies.length, color: "text-indigo-700" },
          { label: "Violations", value: violations, color: violations > 0 ? "text-red-700" : "text-emerald-700" },
          { label: "Kill Switches", value: killSwitches.length, color: "text-slate-700" },
          { label: "Active Kill Switches", value: activeKS, color: activeKS > 0 ? "text-amber-700" : "text-emerald-700" },
        ].map((s) => (
          <div key={s.label} className="bg-white border border-slate-200 rounded-xl p-5 shadow-sm">
            <p className={`text-3xl font-bold ${s.color}`}>{s.value}</p>
            <p className="text-xs text-slate-500 mt-1">{s.label}</p>
          </div>
        ))}
      </div>

      {/* Policy Table */}
      <h3 className="font-semibold text-slate-700 mb-4">Policy Status</h3>
      <div className="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden mb-8">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              {["Policy", "Framework", "Scope", "Status", "Findings", "Last Evaluated"].map((h) => (
                <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wide">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {policies.map((p) => (
              <tr key={p.name} className="hover:bg-slate-50 transition-colors">
                <td className="px-4 py-3 font-medium text-slate-800">{p.name}</td>
                <td className="px-4 py-3">
                  <span className="bg-indigo-100 text-indigo-700 text-xs rounded px-2 py-0.5 font-semibold">
                    {p.framework}
                  </span>
                </td>
                <td className="px-4 py-3 text-slate-500 text-xs">{p.scope}</td>
                <td className="px-4 py-3">
                  <StatusBadge status={p.status} />
                </td>
                <td className="px-4 py-3 text-xs font-semibold">
                  {p.findings > 0 ? (
                    <span className="text-red-700">{p.findings}</span>
                  ) : (
                    <span className="text-slate-400">—</span>
                  )}
                </td>
                <td className="px-4 py-3 text-slate-400 text-xs">{p.lastEvaluated}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Kill Switches */}
      <h3 className="font-semibold text-slate-700 mb-4">Kill Switches</h3>
      <div className="grid grid-cols-2 gap-4 mb-8">
        {killSwitches.map((ks) => (
          <div
            key={ks.id}
            className={`rounded-xl border p-4 flex items-center justify-between ${
              ks.enabled
                ? "bg-red-50 border-red-200"
                : "bg-white border-slate-200"
            }`}
          >
            <div>
              <p className="font-medium text-slate-800">{ks.name}</p>
              <p className="text-xs text-slate-500 mt-0.5">
                Scope: {ks.scope} · Last triggered: {ks.lastTriggered}
              </p>
            </div>
            <div className="flex items-center gap-2">
              <span
                className={`text-xs font-semibold px-2 py-0.5 rounded-full ${
                  ks.enabled
                    ? "bg-red-200 text-red-800"
                    : "bg-slate-100 text-slate-500"
                }`}
              >
                {ks.enabled ? "ACTIVE" : "INACTIVE"}
              </span>
            </div>
          </div>
        ))}
      </div>

      {/* Audit Log */}
      <h3 className="font-semibold text-slate-700 mb-4">
        Audit Log{" "}
        <span className="text-slate-400 font-normal text-sm">(50 entries today)</span>
      </h3>
      <div className="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              {["Time", "Actor", "Action", "Resource", "Outcome"].map((h) => (
                <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wide">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {auditLog.map((entry, i) => (
              <tr key={i} className="hover:bg-slate-50 transition-colors">
                <td className="px-4 py-3 font-mono text-xs text-slate-500">{entry.time}</td>
                <td className="px-4 py-3 font-mono text-xs text-indigo-700">{entry.actor}</td>
                <td className="px-4 py-3 font-mono text-xs text-slate-700">{entry.action}</td>
                <td className="px-4 py-3 text-xs text-slate-500">{entry.resource}</td>
                <td className="px-4 py-3">
                  <span className={`text-xs px-2 py-0.5 rounded font-semibold ${outcomeColors[entry.outcome]}`}>
                    {entry.outcome.toUpperCase()}
                  </span>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
