"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";

const navItems = [
  { href: "/", label: "Dashboard", icon: "⬡" },
  { href: "/dashboard", label: "Debt Heatmap", icon: "🔥" },
  { href: "/certs", label: "Cert Health", icon: "🔒" },
  { href: "/agents", label: "Agent Composer", icon: "🤖" },
  { href: "/governance", label: "Governance", icon: "🛡️" },
  { href: "/sustainability", label: "Sustainability", icon: "🌱" },
];

export default function Sidebar() {
  const pathname = usePathname();

  return (
    <aside className="w-64 bg-indigo-900 text-white flex flex-col shrink-0">
      <div className="px-6 py-5 border-b border-indigo-700">
        <span className="text-xs font-semibold uppercase tracking-widest text-indigo-300">
          Control Plane
        </span>
        <h1 className="text-xl font-bold mt-1">Modernization Hub</h1>
      </div>

      <nav className="flex-1 px-3 py-4 space-y-1">
        {navItems.map(({ href, label, icon }) => {
          const active = pathname === href;
          return (
            <Link
              key={href}
              href={href}
              className={`flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm font-medium transition-colors ${
                active
                  ? "bg-indigo-700 text-white"
                  : "text-indigo-200 hover:bg-indigo-800 hover:text-white"
              }`}
            >
              <span className="text-base">{icon}</span>
              {label}
            </Link>
          );
        })}
      </nav>

      <div className="px-6 py-4 border-t border-indigo-700 text-xs text-indigo-400">
        Module 10 · v1.0.0
      </div>
    </aside>
  );
}
