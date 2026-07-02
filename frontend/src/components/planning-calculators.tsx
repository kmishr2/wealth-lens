"use client";

import { useActionState } from "react";
import { calculateSIPAction, compareWhatIfAction, type SIPState, type WhatIfState } from "@/app/actions/projections";

const sipInitial: SIPState = { message: "" };
const whatIfInitial: WhatIfState = { message: "" };

export function SIPCalculator({ portfolioID, currency }: { portfolioID: string; currency: string }) {
  const [state, action, pending] = useActionState(calculateSIPAction, sipInitial);
  return (
    <form action={action} className="rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6">
      <input name="portfolioId" type="hidden" value={portfolioID} /><input name="currency" type="hidden" value={currency} />
      <p className="eyebrow">Single scenario</p><h2 className="mt-2 text-2xl font-semibold">SIP projection</h2>
      <div className="mt-5 grid gap-4 sm:grid-cols-2"><ProjectionFields /></div>
      <button className="focus-ring mt-5 rounded-xl bg-[var(--brand)] px-5 py-3 font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Calculating…" : "Calculate"}</button>
      {state.data ? (
        <div className="mt-6 grid gap-3 border-t border-[var(--line)] pt-5 sm:grid-cols-3">
          <Result label="Projected value" value={money(state.data.projected_nominal_value, currency)} />
          <Result label="Inflation-adjusted" value={money(state.data.projected_real_value, currency)} />
          <Result label="Contributions" value={money(state.data.total_contributions, currency)} />
        </div>
      ) : state.message ? <Message text={state.message} /> : null}
    </form>
  );
}

export function WhatIfCalculator({ portfolioID, currency }: { portfolioID: string; currency: string }) {
  const [state, action, pending] = useActionState(compareWhatIfAction, whatIfInitial);
  return (
    <form action={action} className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6">
      <input name="portfolioId" type="hidden" value={portfolioID} /><input name="currency" type="hidden" value={currency} />
      <p className="eyebrow">Compare assumptions</p><h2 className="mt-2 text-2xl font-semibold">What-if scenarios</h2>
      <div className="mt-5 grid gap-6 lg:grid-cols-2">
        <Scenario title="Baseline"><ProjectionFields prefix="baseline" /></Scenario>
        <Scenario title="Alternative"><ProjectionFields prefix="alternative" /></Scenario>
      </div>
      <button className="focus-ring mt-5 rounded-xl bg-[var(--brand)] px-5 py-3 font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Comparing…" : "Compare scenarios"}</button>
      {state.data ? (
        <div className="mt-6 border-t border-[var(--line)] pt-5">
          <p className="text-sm font-semibold">Alternative difference from baseline</p>
          <div className="mt-3 grid gap-3 sm:grid-cols-3">
            <Result label="Nominal value" value={signedMoney(state.data.scenarios[1].nominal_difference_from_baseline, currency)} />
            <Result label="Real value" value={signedMoney(state.data.scenarios[1].real_difference_from_baseline, currency)} />
            <Result label="Contributions" value={signedMoney(state.data.scenarios[1].contribution_difference_from_baseline, currency)} />
          </div>
        </div>
      ) : state.message ? <Message text={state.message} /> : null}
    </form>
  );
}

function ProjectionFields({ prefix = "" }: { prefix?: string }) {
  const name = (field: string) => prefix ? `${prefix}${field[0].toUpperCase()}${field.slice(1)}` : field;
  return <>
    <Input label="Initial investment" name={name("initialInvestment")} defaultValue="0" min="0" />
    <Input label="Monthly contribution" name={name("monthlyContribution")} min="0" />
    <Input label="Annual return %" name={name("annualReturn")} min="-99.99" />
    <Input label="Inflation %" name={name("inflation")} defaultValue="0" min="0" />
    <Input label="Months" name={name("months")} step="1" min="1" />
  </>;
}

function Input({ label, name, defaultValue, step = "any", min }: { label: string; name: string; defaultValue?: string; step?: string; min: string }) {
  return <label className="text-sm font-semibold">{label}<input className="planning-input" defaultValue={defaultValue} min={min} name={name} step={step} type="number" /></label>;
}
function Scenario({ title, children }: { title: string; children: React.ReactNode }) { return <fieldset className="grid gap-4 rounded-2xl border border-[var(--line)] p-4 sm:grid-cols-2"><legend className="px-2 font-semibold">{title}</legend>{children}</fieldset>; }
function Result({ label, value }: { label: string; value: string }) { return <div className="rounded-xl bg-[var(--brand-soft)] p-4"><p className="text-xs text-[var(--muted)]">{label}</p><p className="mt-1 font-mono font-semibold">{value}</p></div>; }
function Message({ text }: { text: string }) { return <p className="mt-5 rounded-xl border border-[#e8c9c4] bg-[#fff4f2] px-4 py-3 text-sm text-[var(--danger)]">{text}</p>; }
function money(value: string, currency: string) { return new Intl.NumberFormat("en-IN", { style: "currency", currency, maximumFractionDigits: 2 }).format(Number(value)); }
function signedMoney(value: string, currency: string) { const amount = Number(value); return `${amount > 0 ? "+" : ""}${money(value, currency)}`; }
