type Variant = "healthy" | "warning" | "critical" | "expired" | "active" | "idle" | "compliant" | "violation";

const styles: Record<Variant, string> = {
  healthy: "bg-emerald-100 text-emerald-800",
  warning: "bg-amber-100 text-amber-800",
  critical: "bg-red-100 text-red-800",
  expired: "bg-red-200 text-red-900",
  active: "bg-blue-100 text-blue-800",
  idle: "bg-slate-100 text-slate-600",
  compliant: "bg-emerald-100 text-emerald-800",
  violation: "bg-red-100 text-red-800",
};

export default function StatusBadge({ status }: { status: Variant }) {
  return (
    <span
      className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-semibold capitalize ${styles[status]}`}
    >
      {status}
    </span>
  );
}
