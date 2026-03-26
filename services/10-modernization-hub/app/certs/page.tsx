import StatusBadge from "@/components/StatusBadge";

type CertStatus = "healthy" | "warning" | "critical" | "expired";

interface Cert {
  cn: string;
  owner: string;
  ca: string;
  expiry: string;
  daysLeft: number;
  status: CertStatus;
  autoRenew: boolean;
}

const certs: Cert[] = [
  { cn: "api.payments.internal", owner: "Payments Team", ca: "Internal CA", expiry: "2026-01-15", daysLeft: 180, status: "healthy", autoRenew: true },
  { cn: "auth.platform.io", owner: "Identity Team", ca: "Let's Encrypt", expiry: "2025-12-01", daysLeft: 135, status: "healthy", autoRenew: true },
  { cn: "inventory.svc.internal", owner: "Ops Team", ca: "Internal CA", expiry: "2025-11-20", daysLeft: 124, status: "healthy", autoRenew: false },
  { cn: "reporting.data.io", owner: "Analytics", ca: "DigiCert", expiry: "2025-10-05", daysLeft: 78, status: "healthy", autoRenew: true },
  { cn: "notification.bus.internal", owner: "Platform", ca: "Internal CA", expiry: "2025-09-18", daysLeft: 61, status: "healthy", autoRenew: false },
  { cn: "user-profile.api.io", owner: "Product", ca: "Let's Encrypt", expiry: "2025-09-01", daysLeft: 44, status: "healthy", autoRenew: true },
  { cn: "gateway.edge.io", owner: "Platform", ca: "Cloudflare", expiry: "2025-08-28", daysLeft: 40, status: "healthy", autoRenew: true },
  { cn: "admin.dashboard.internal", owner: "SRE", ca: "Internal CA", expiry: "2025-08-15", daysLeft: 27, status: "warning", autoRenew: false },
  { cn: "billing.api.internal", owner: "Finance", ca: "Internal CA", expiry: "2025-08-10", daysLeft: 22, status: "warning", autoRenew: false },
  { cn: "legacy.sso.io", owner: "Identity Team", ca: "Comodo", expiry: "2025-08-05", daysLeft: 17, status: "warning", autoRenew: false },
  { cn: "old-reporting.internal", owner: "Analytics", ca: "Self-Signed", expiry: "2025-07-25", daysLeft: 6, status: "critical", autoRenew: false },
  { cn: "devportal.io", owner: "Dev Exp", ca: "Let's Encrypt", expiry: "2025-07-20", daysLeft: 1, status: "critical", autoRenew: false },
  { cn: "staging-api.internal", owner: "QA Team", ca: "Self-Signed", expiry: "2025-07-10", daysLeft: -9, status: "expired", autoRenew: false },
  { cn: "test-auth.legacy.io", owner: "Identity Team", ca: "Self-Signed", expiry: "2025-06-30", daysLeft: -19, status: "expired", autoRenew: false },
];

const fill = 14;
const healthy = certs.filter((c) => c.status === "healthy").length;
const warning = certs.filter((c) => c.status === "warning").length;
const critical = certs.filter((c) => c.status === "critical").length;
const expired = certs.filter((c) => c.status === "expired").length;

function DaysCell({ days }: { days: number }) {
  const color =
    days < 0
      ? "text-red-700 font-bold"
      : days <= 7
        ? "text-red-600 font-semibold"
        : days <= 30
          ? "text-amber-600 font-semibold"
          : "text-slate-700";
  return (
    <span className={color}>
      {days < 0 ? `${Math.abs(days)} days ago` : `${days} days`}
    </span>
  );
}

export default function CertsPage() {
  return (
    <div>
      <div className="mb-8">
        <h2 className="text-2xl font-bold text-slate-800">
          Certificate Health Dashboard
        </h2>
        <p className="text-slate-500 mt-1">
          Expiry tracking, renewal status, and compliance across all discovered certs
        </p>
      </div>

      {/* Summary */}
      <div className="grid grid-cols-4 gap-4 mb-8">
        {[
          { label: "Healthy", count: healthy, border: "border-emerald-200", bg: "bg-emerald-50", text: "text-emerald-700" },
          { label: "Warning (≤30d)", count: warning, border: "border-amber-200", bg: "bg-amber-50", text: "text-amber-700" },
          { label: "Critical (≤7d)", count: critical, border: "border-red-200", bg: "bg-red-50", text: "text-red-700" },
          { label: "Expired", count: expired, border: "border-red-300", bg: "bg-red-100", text: "text-red-900" },
        ].map((item) => (
          <div key={item.label} className={`rounded-xl border ${item.border} ${item.bg} p-5`}>
            <p className={`text-3xl font-bold ${item.text}`}>{item.count}</p>
            <p className="text-xs text-slate-500 mt-1">{item.label}</p>
          </div>
        ))}
      </div>

      {/* Compliance notice */}
      <div className="mb-6 flex gap-3">
        <div className="flex-1 rounded-lg bg-indigo-50 border border-indigo-200 px-4 py-3 text-sm text-indigo-800">
          <strong>200-day validity threshold:</strong> {certs.filter((c) => c.daysLeft > 0 && c.daysLeft <= 200).length} certs within compliance window
        </div>
        <div className="flex-1 rounded-lg bg-amber-50 border border-amber-200 px-4 py-3 text-sm text-amber-800">
          <strong>47-day CA/B Forum threshold:</strong> {certs.filter((c) => c.daysLeft > 0 && c.daysLeft <= 47).length} certs require immediate action
        </div>
      </div>

      {/* Table */}
      <div className="bg-white border border-slate-200 rounded-xl shadow-sm overflow-hidden">
        <table className="w-full text-sm">
          <thead className="bg-slate-50 border-b border-slate-200">
            <tr>
              {["Common Name", "Owner", "CA", "Expiry Date", "Days Left", "Status", "Auto-Renew"].map((h) => (
                <th key={h} className="px-4 py-3 text-left text-xs font-semibold text-slate-500 uppercase tracking-wide">
                  {h}
                </th>
              ))}
            </tr>
          </thead>
          <tbody className="divide-y divide-slate-50">
            {certs.map((cert) => (
              <tr key={cert.cn} className="hover:bg-slate-50 transition-colors">
                <td className="px-4 py-3 font-mono text-indigo-700 text-xs">{cert.cn}</td>
                <td className="px-4 py-3 text-slate-600">{cert.owner}</td>
                <td className="px-4 py-3 text-slate-500 text-xs">{cert.ca}</td>
                <td className="px-4 py-3 text-slate-600 text-xs">{cert.expiry}</td>
                <td className="px-4 py-3 text-xs">
                  <DaysCell days={cert.daysLeft} />
                </td>
                <td className="px-4 py-3">
                  <StatusBadge status={cert.status} />
                </td>
                <td className="px-4 py-3">
                  {cert.autoRenew ? (
                    <span className="inline-flex items-center gap-1 text-xs text-emerald-700 font-medium">
                      ✓ Enrolled
                    </span>
                  ) : (
                    <button className="text-xs text-indigo-600 hover:text-indigo-800 font-medium underline underline-offset-2">
                      Enroll →
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
