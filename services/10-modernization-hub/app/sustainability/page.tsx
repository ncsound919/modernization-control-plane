interface KpiMetric {
  service: string;
  team: string;
  gco2ePerCall: number;
  callsToday: number;
  totalGco2e: number;
  region: string;
  greenRegion: boolean;
}

const metrics: KpiMetric[] = [
  { service: "api-gateway", team: "Platform", gco2ePerCall: 0.12, callsToday: 125000, totalGco2e: 15.0, region: "us-west-2", greenRegion: true },
  { service: "auth.platform.io", team: "Identity", gco2ePerCall: 0.08, callsToday: 98000, totalGco2e: 7.8, region: "eu-west-1", greenRegion: true },
  { service: "billing-monolith", team: "Finance", gco2ePerCall: 0.55, callsToday: 42000, totalGco2e: 23.1, region: "us-east-1", greenRegion: false },
  { service: "user-profile-api", team: "Product", gco2ePerCall: 0.09, callsToday: 210000, totalGco2e: 18.9, region: "us-west-2", greenRegion: true },
  { service: "reporting-svc", team: "Analytics", gco2ePerCall: 0.31, callsToday: 15000, totalGco2e: 4.65, region: "us-east-1", greenRegion: false },
  { service: "inventory-v1", team: "Ops", gco2ePerCall: 0.22, callsToday: 31000, totalGco2e: 6.82, region: "ap-southeast-1", greenRegion: false },
  { service: "notification-bus", team: "Platform", gco2ePerCall: 0.06, callsToday: 180000, totalGco2e: 10.8, region: "eu-west-1", greenRegion: true },
  { service: "legacy-auth-service", team: "Identity", gco2ePerCall: 0.74, callsToday: 18000, totalGco2e: 13.3, region: "us-east-1", greenRegion: false },
  { service: "cloud-native-auth", team: "Identity", gco2ePerCall: 0.07, callsToday: 21000, totalGco2e: 1.47, region: "eu-west-1", greenRegion: true },
  { service: "svc-inventory-v2", team: "Ops", gco2ePerCall: 0.10, callsToday: 29000, totalGco2e: 2.9, region: "us-west-2", greenRegion: true },
];

const totalGco2e = metrics.reduce((sum, m) => sum + m.totalGco2e, 0);
const baseline = 162.5;
const reduction = Math.round((1 - totalGco2e / baseline) * 100);
const greenCallPct = Math.round(
  (metrics.filter((m) => m.greenRegion).reduce((s, m) => s + m.callsToday, 0) /
    metrics.reduce((s, m) => s + m.callsToday, 0)) *
    100
);

const carbonBudget = 150;
const budgetUsedPct = Math.round((totalGco2e / carbonBudget) * 100);

export default function SustainabilityPage() {
  return (
    <div>
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-slate-800">
          Sustainability KPI Board
        </h2>
        <p className="text-slate-500 mt-1">
          gCO₂e emissions, carbon budget, and green region utilization
        </p>
      </div>

      {/* Top KPIs */}
      <div className="grid grid-cols-2 gap-4 mb-8 xl:grid-cols-4">
        {[
          { label: "Total gCO₂e Today", value: `${totalGco2e.toFixed(1)} g`, sub: "across all services", color: "text-teal-700" },
          { label: "vs. Baseline", value: `−${reduction}%`, sub: `Baseline: ${baseline} g/day`, color: "text-emerald-700" },
          { label: "Green Region Calls", value: `${greenCallPct}%`, sub: "ISO 14001 certified regions", color: "text-emerald-700" },
          { label: "Carbon Budget Used", value: `${budgetUsedPct}%`, sub: `Budget: ${carbonBudget} gCO₂e/day`, color: budgetUsedPct > 90 ? "text-red-700" : "text-teal-700" },
        ].map((kpi) => (
          <div key={kpi.label} className="bg-white border border-slate-200 rounded-xl p-5 shadow-sm">
            <p className={`text-3xl font-bold ${kpi.color}`}>{kpi.value}</p>
            <p className="text-sm font-medium text-slate-700 mt-1">{kpi.label}</p>
            <p className="text-xs text-slate-400 mt-0.5">{kpi.sub}</p>
          </div>
        ))}
      </div>

      {/* Carbon budget bar */}
      <div className="bg-white border border-slate-200 rounded-xl shadow-sm p-6 mb-8">
        <div className="flex items-center justify-between mb-2">
          <h3 className="font-semibold text-slate-700">Monthly Carbon Budget</h3>
          <span className="text-sm text-slate-500">
            {totalGco2e.toFixed(1)} / {carbonBudget} gCO₂e ({budgetUsedPct}% used)
          </span>
        </div>
        <div className="h-4 bg-slate-100 rounded-full overflow-hidden">
          <div
            className={`h-full rounded-full transition-all ${
              budgetUsedPct > 90
                ? "bg-red-500"
                : budgetUsedPct > 70
                  ? "bg-amber-400"
                  : "bg-emerald-500"
            }`}
            style={{ width: `${Math.min(budgetUsedPct, 100)}%` }}
          />
        </div>
        <div className="flex justify-between mt-1 text-xs text-slate-400">
          <span>0 g</span>
          <span className="text-amber-600">70% threshold</span>
          <span className="text-red-600">90% alert</span>
          <span>{carbonBudget} g</span>
        </div>
      </div>

      {/* Per-service table */}
      <h3 className="font-semibold text-slate-700 mb-4">Per-Service Emissions</h3>
      <div className="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              {["Service", "Team", "gCO₂e / Call", "Calls Today", "Total gCO₂e", "Region", "Green"].map((h) => (
                <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wide">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {metrics
              .slice()
              .sort((a, b) => b.totalGco2e - a.totalGco2e)
              .map((m) => (
                <tr key={m.service} className="hover:bg-slate-50 transition-colors">
                  <td className="px-4 py-3 font-mono text-indigo-700 text-xs">{m.service}</td>
                  <td className="px-4 py-3 text-slate-600">{m.team}</td>
                  <td className="px-4 py-3 text-slate-700 font-semibold">{m.gco2ePerCall.toFixed(2)}</td>
                  <td className="px-4 py-3 text-slate-600">{m.callsToday.toLocaleString()}</td>
                  <td className="px-4 py-3">
                    <div className="flex items-center gap-2">
                      <div className="w-20 h-1.5 bg-slate-100 rounded-full overflow-hidden">
                        <div
                          className="h-full bg-teal-500 rounded-full"
                          style={{ width: `${Math.min((m.totalGco2e / 25) * 100, 100)}%` }}
                        />
                      </div>
                      <span className="text-xs font-semibold text-slate-700">
                        {m.totalGco2e.toFixed(2)} g
                      </span>
                    </div>
                  </td>
                  <td className="px-4 py-3 font-mono text-xs text-slate-500">{m.region}</td>
                  <td className="px-4 py-3">
                    {m.greenRegion ? (
                      <span className="text-emerald-600 font-semibold text-xs">✓ Green</span>
                    ) : (
                      <span className="text-slate-400 text-xs">—</span>
                    )}
                  </td>
                </tr>
              ))}
          </tbody>
          <tfoot className="bg-slate-50 border-t border-slate-200">
            <tr>
              <td colSpan={4} className="px-4 py-3 text-xs font-semibold text-slate-600 uppercase">
                Total
              </td>
              <td className="px-4 py-3 font-bold text-teal-700">
                {totalGco2e.toFixed(2)} g
              </td>
              <td colSpan={2} />
            </tr>
          </tfoot>
        </table>
      </div>

      {/* ESG badge */}
      <div className="mt-6 rounded-xl bg-emerald-50 border border-emerald-200 p-4 flex items-center gap-3">
        <span className="text-2xl">🌱</span>
        <div>
          <p className="font-semibold text-emerald-800">ESG Report Ready</p>
          <p className="text-xs text-emerald-700 mt-0.5">
            Q3 2025 carbon report available for download — {reduction}% reduction validated against ISO 14064-1 baseline
          </p>
        </div>
        <button className="ml-auto text-xs bg-emerald-700 text-white rounded-lg px-4 py-2 hover:bg-emerald-800 transition-colors font-medium">
          Export Report
        </button>
      </div>
    </div>
  );
}
