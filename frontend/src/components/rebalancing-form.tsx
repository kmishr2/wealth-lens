"use client";

import { useActionState } from "react";
import { calculateRebalancingAction, type RebalancingState } from "@/app/actions/rebalancing";
import type { AssetClassAllocation } from "@/lib/types";

const initialState: RebalancingState = { message: "" };

export function RebalancingForm({ portfolioID, allocations }: { portfolioID: string; allocations: AssetClassAllocation[] }) {
  const [state, action, pending] = useActionState(calculateRebalancingAction, initialState);
  return (
    <form action={action} className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6">
      <input name="portfolioId" type="hidden" value={portfolioID} />
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div><p className="eyebrow">Explicit targets</p><h2 className="mt-2 text-2xl font-semibold">Rebalancing</h2></div>
        <label className="text-sm font-semibold">Drift tolerance %<input className="planning-input w-32!" defaultValue="5" min="0" name="tolerance" step="0.1" type="number" /></label>
      </div>
      {allocations.length === 0 ? <p className="mt-5 text-sm text-[var(--muted)]">A complete current allocation is required.</p> : (
        <div className="mt-5 space-y-3">
          {allocations.map((item) => (
            <label className="grid grid-cols-[minmax(0,1fr)_110px] items-center gap-4 rounded-xl border border-[var(--line)] p-4" key={`${item.currency}-${item.asset_class}`}>
              <span><strong className="capitalize">{item.asset_class.replaceAll("_", " ")}</strong><span className="ml-2 text-xs text-[var(--muted)]">{item.currency} · current {percent(item.percentage)}</span></span>
              <input className="planning-input mt-0!" defaultValue={item.percentage} max="100" min="0" name={`target::${item.currency}::${item.asset_class}`} step="0.01" type="number" />
            </label>
          ))}
        </div>
      )}
      <button className="focus-ring mt-5 rounded-xl bg-[var(--brand)] px-5 py-3 font-semibold text-white disabled:opacity-60" disabled={pending || allocations.length === 0} type="submit">{pending ? "Calculating…" : "Calculate drift"}</button>
      {state.data ? (
        <div className="mt-6 overflow-x-auto border-t border-[var(--line)] pt-5">
          <table className="w-full min-w-160 text-left text-sm"><thead className="text-xs uppercase tracking-wide text-[var(--muted)]"><tr><th className="pb-3">Class</th><th className="pb-3">Current</th><th className="pb-3">Target</th><th className="pb-3">Drift</th><th className="pb-3">Adjustment</th></tr></thead>
            <tbody className="divide-y divide-[var(--line)]">{state.data.items.map((item) => <tr key={`${item.currency}-${item.asset_class}`}><td className="py-3 font-semibold capitalize">{item.asset_class} <span className="text-xs text-[var(--muted)]">{item.currency}</span></td><td>{percent(item.current_percentage)}</td><td>{percent(item.target_percentage)}</td><td className={item.is_outside_tolerance ? "text-[var(--danger)]" : "text-[var(--muted)]"}>{percent(item.drift_percentage)}</td><td className="font-mono">{item.action === "none" ? "Within tolerance" : `${item.action} ${money(item.suggested_adjustment, item.currency)}`}</td></tr>)}</tbody>
          </table>
        </div>
      ) : state.message ? <p className="mt-5 rounded-xl border border-[#e8c9c4] bg-[#fff4f2] px-4 py-3 text-sm text-[var(--danger)]">{state.message}</p> : null}
    </form>
  );
}

function percent(value: string) { return `${new Intl.NumberFormat("en-IN", { maximumFractionDigits: 2 }).format(Number(value))}%`; }
function money(value: string, currency: string) { return new Intl.NumberFormat("en-IN", { style: "currency", currency, maximumFractionDigits: 2 }).format(Math.abs(Number(value))); }
