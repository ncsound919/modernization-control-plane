import Link from "next/link";

const summaryCards = [
  {
    label: "Total Assets",
    value: "247",
    change: "+12 this week",
    color: "bg-indigo-600",
    icon: "📦",
    href: "/dashboard",
  },
  {
    label: "Active Certs",
    value: "89",
    change: "5 expiring soon",
    color: "bg-blue-600",
    icon: "🔒",
    href: "/certs",
  },
  {
    label: "Running Agents",
    value: "12",
    change: "4 workflows active",
    color: "bg-violet-600",
    icon: "🤖",
    href: "/agents",
  },
  {
    label: "Compliance Score",
    value: "94%",
    change: "+2% vs last month",
    color: "bg-emerald-600",
    icon: "🛡️",
    href: "/governance",
  },
];

const recentActivity = [
  {
    time: "2 min ago",
    event: "Certificate auto-renewed",
    detail: "api.payments.internal — 365 days added",
    type: "cert",
  },
  {
    time: "8 min ago",
    event: "Migration agent completed",
    detail: "legacy-auth-service → cloud-native-auth",
    type: "agent",
  },
  {
    time: "15 min ago",
    event: "Kill switch activated",
    detail: "Tenant: acme-corp — workflow paused",
    type: "governance",
  },
  {
    time: "31 min ago",
    event: "Debt score updated",
    detail: "billing-monolith risk: HIGH → MEDIUM",
    type: "debt",
  },
  {
    time: "1 hr ago",
    event: "New asset discovered",
    detail: "svc-inventory-v2 registered in Discovery Engine",
    type: "discovery",
  },
  {
    time: "2 hr ago",
    event: "GDPR policy evaluated",
    detail: "12 services — all compliant",
    type: "governance",
  },
  {
    time: "3 hr ago",
    event: "Carbon budget alert cleared",
    detail: "Monthly gCO2e within 10% of target",
    type: "sustainability",
  },
];

const typeColors: Record<string, string> = {
  cert: "bg-blue-500",
  agent: "bg-violet-500",
  governance: "bg-indigo-500",
  debt: "bg-amber-500",
  discovery: "bg-emerald-500",
  sustainability: "bg-teal-500",
};

export default function HomePage() {
  return (
    <div>
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-slate-800">
          Platform Overview
        </h2>
        <p className="text-slate-500 mt-1">
          Real-time snapshot across all Modernization Control Plane modules
        </p>
      </div>

      {/* Summary cards */}
      <div className="grid grid-cols-1 gap-5 sm:grid-cols-2 xl:grid-cols-4 mb-10">
        {summaryCards.map((card) => (
          <Link
            key={card.label}
            href={card.href}
            className="group rounded-xl bg-white border border-slate-200 shadow-sm hover:shadow-md transition-shadow p-6"
          >
            <div className="flex items-start justify-between">
              <div>
                <p className="text-sm text-slate-500 font-medium">
                  {card.label}
                </p>
                <p className="mt-1 text-3xl font-bold text-slate-800">
                  {card.value}
                </p>
                <p className="mt-1 text-xs text-slate-400">{card.change}</p>
              </div>
              <span
                className={`${card.color} text-white rounded-lg p-2.5 text-xl`}
              >
                {card.icon}
              </span>
            </div>
          </Link>
        ))}
      </div>

      {/* Recent activity */}
      <div className="bg-white border border-slate-200 rounded-xl shadow-sm">
        <div className="px-6 py-4 border-b border-slate-100 flex items-center justify-between">
          <h3 className="font-semibold text-slate-800">Recent Activity</h3>
          <span className="text-xs text-slate-400 bg-slate-100 rounded-full px-3 py-1">
            Live feed
          </span>
        </div>
        <ul className="divide-y divide-slate-50">
          {recentActivity.map((item, i) => (
            <li key={i} className="px-6 py-4 flex items-start gap-4">
              <span
                className={`mt-0.5 w-2.5 h-2.5 rounded-full shrink-0 ${typeColors[item.type] ?? "bg-slate-400"}`}
              />
              <div className="flex-1 min-w-0">
                <p className="text-sm font-medium text-slate-800">
                  {item.event}
                </p>
                <p className="text-xs text-slate-500 truncate">{item.detail}</p>
              </div>
              <span className="text-xs text-slate-400 whitespace-nowrap">
                {item.time}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
