"use client";

import { useActionState } from "react";
import { recordFixedDepositValueAction } from "@/app/actions/fixed-deposits";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function FixedDepositValueForm({ portfolioID, accountID, fixedDepositID, currency }: { portfolioID: string; accountID: string; fixedDepositID: string; currency: string }) {
  const [state, action, pending] = useActionState(recordFixedDepositValueAction, initialState);
  return (
    <details className="mt-5 border-t border-[var(--line)] pt-4">
      <summary className="cursor-pointer text-sm font-semibold text-[var(--brand)]">Record updated value</summary>
      <form action={action} className="mt-4 grid gap-3" noValidate>
        <input name="portfolioId" type="hidden" value={portfolioID} />
        <input name="accountId" type="hidden" value={accountID} />
        <input name="fixedDepositId" type="hidden" value={fixedDepositID} />
        <Field label={`Current value (${currency})`} error={state.fields?.currentValue}><input className="planning-input" min="0" name="currentValue" step="0.01" type="number" /></Field>
        <Field label="Value date" error={state.fields?.currentValueDate}><input className="planning-input" name="currentValueDate" type="date" /></Field>
        {state.message && <p aria-live="polite" className={`rounded-xl px-3 py-2 text-xs ${state.success ? "bg-[#edf7f1] text-[var(--brand)]" : "bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>}
        <button className="focus-ring justify-self-start rounded-xl bg-[var(--brand)] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Recording…" : "Record value"}</button>
      </form>
    </details>
  );
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>;
}
