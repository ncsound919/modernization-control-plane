import StatusBadge from "@/components/StatusBadge";

type RiskLevel = "critical" | "warning" | "healthy";

interface Service {
  name: string;
  team: string;
  domain: string;
  riskScore: number;
  risk: RiskLevel;
  debtItems: number;
  readiness: number;
  lastScanned: string;
}

const services: Service[] = [
  {
    name: "billing-monolith",
    team: "Payments",
    domain: "Finance",
    riskScore: 92,
    risk: "critical",
    debtItems: 38,
    readiness: 12,
    lastScanned: "2 hr ago",
  },
  {
    name: "legacy-auth-service",
    team: "Identity",
    domain: "Security",
    riskScore: 87,
    risk: "critical",
    debtItems: 31,
    readiness: 25,
    lastScanned: "1 hr ago",
  },
  {
    name: "inventory-v1",
    team: "Ops",
    domain: "Supply Chain",
    riskScore: 74,
    risk: "warning",
    debtItems: 22,
    readiness: 41,
    lastScanned: "3 hr ago",
  },
  {
    name: "reporting-svc",
    team: "Analytics",
    domain: "Data",
    riskScore: 68,
    risk: "warning",
    debtItems: 19,
    readiness: 55,
    lastScanned: "30 min ago",
  },
  {
    name: "notification-bus",
    team: "Platform",
    domain: "Messaging",
    riskScore: 61,
    risk: "warning",
    debtItems: 14,
    readiness: 60,
    lastScanned: "45 min ago",
  },
  {
    name: "user-profile-api",
    team: "Product",
    domain: "Customer",
    riskScore: 45,
    risk: "warning",
    debtItems: 9,
    readiness: 72,
    lastScanned: "1 hr ago",
  },
  {
    name: "api-gateway",
    team: "Platform",
    domain: "Infrastructure",
    riskScore: 28,
    risk: "healthy",
    debtItems: 4,
    readiness: 88,
    lastScanned: "15 min ago",
  },
  {
    name: "cloud-native-auth",
    team: "Identity",
    domain: "Security",
    riskScore: 18,
    risk: "healthy",
    debtItems: 2,
    readiness: 95,
    lastScanned: "5 min ago",
  },
  {
    name: "svc-inventory-v2",
    team: "Ops",
    domain: "Supply Chain",
    riskScore: 14,
    risk: "healthy",
    debtItems: 1,
    readiness: 97,
    lastScanned: "20 min ago",
  },
];

const riskColor: Record<RiskLevel, string> = {
  critical: "bg-red-500",
  warning: "bg-amber-400",
  healthy: "bg-emerald-500",
};

function RiskBar({ score }: { score: number }) {
  const color =
    score >= 75 ? "bg-red-500" : score >= 50 ? "bg-amber-400" : "bg-emerald-500";
  return (
    <div className="flex items-center gap-2">
      <div className="flex-1 h-2 bg-slate-100 rounded-full overflow-hidden">
        <div className={`h-full rounded-full ${color}`} style={{ width: `${score}%` }} />
      </div>
      <span className="text-xs font-semibold w-7 text-right">{score}</span>
    </div>
  );
}

export default function DebtHeatmapPage() {
  const critical = services.filter((s) => s.risk === "critical").length;
  const warning = services.filter((s) => s.risk === "warning").length;
  const healthy = services.filter((s) => s.risk === "healthy").length;

  return (
    <div>
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-slate-800">Debt Heatmap</h2>
        <p className="text-slate-500 mt-1">
          Technical debt risk scores across all registered services
        </p>
      </div>

      {/* Legend summary */}
      <div className="grid grid-cols-3 gap-4 mb-8">
        {[
          { label: "Critical", count: critical, color: "bg-red-500", text: "text-red-700", bg: "bg-red-50 border-red-200" },
          { label: "Warning", count: warning, color: "bg-amber-400", text: "text-amber-700", bg: "bg-amber-50 border-amber-200" },
          { label: "Healthy", count: healthy, color: "bg-emerald-500", text: "text-emerald-700", bg: "bg-emerald-50 border-emerald-200" },
        ].map((item) => (
          <div key={item.label} className={`rounded-xl border ${item.bg} p-4 flex items-center gap-3`}>
            <span className={`w-3 h-3 rounded-full ${item.color}`} />
            <div>
              <p className={`text-2xl font-bold ${item.text}`}>{item.count}</p>
              <p className="text-xs text-slate-500">{item.label} services</p>
            </div>
          </div>
        ))}
      </div>

      {/* Services table */}
      <div className="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              {["Service", "Team", "Domain", "Risk Score", "Status", "Debt Items", "Readiness %", "Last Scanned"].map((h) => (
                <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wide">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {services.map((svc) => (
              <tr key={svc.name} className="hover:bg-slate-50 transition-colors">
                <td className="px-4 py-3 font-mono font-medium text-indigo-700">{svc.name}</td>
                <td className="px-4 py-3 text-slate-600">{svc.team}</td>
                <td className="px-4 py-3 text-slate-500">{svc.domain}</td>
                <td className="px-4 py-3 w-36">
                  <RiskBar score={svc.riskScore} />
                </td>
                <td className="px-4 py-3">
                  <StatusBadge status={svc.risk} />
                </td>
                <td className="px-4 py-3 text-slate-700 font-semibold">{svc.debtItems}</td>
                <td className="px-4 py-3">
                  <div className="flex items-center gap-2">
                    <div className="flex-1 h-2 bg-slate-100 rounded-full overflow-hidden">
                      <div
                        className="h-full bg-indigo-500 rounded-full"
                        style={{ width: `${svc.readiness}%` }}
                      />
                    </div>
                    <span className="text-xs w-7">{svc.readiness}%</span>
                  </div>
                </td>
                <td className="px-4 py-3 text-slate-400 text-xs">{svc.lastScanned}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {/* Heatmap grid */}
      <div className="mt-8">
        <h3 className="font-semibold text-slate-700 mb-4">Visual Heatmap</h3>
        <div className="grid grid-cols-3 gap-3">
          {services.map((svc) => (
            <div
              key={svc.name}
              className={`rounded-lg p-4 text-white ${riskColor[svc.risk]}`}
              style={{ opacity: 0.4 + (svc.riskScore / 100) * 0.6 }}
            >
              <p className="font-semibold text-sm">{svc.name}</p>
              <p className="text-xs mt-1 opacity-90">{svc.team} · Score {svc.riskScore}</p>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
}
