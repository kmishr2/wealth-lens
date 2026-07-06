"use client";

import { useActionState } from "react";
import { closeFixedDepositAction } from "@/app/actions/fixed-deposits";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function FixedDepositCloseForm({ portfolioID, accountID, fixedDepositID, currency }: { portfolioID: string; accountID: string; fixedDepositID: string; currency: string }) {
  const [state, action, pending] = useActionState(closeFixedDepositAction, initialState);
  return (
    <details className="mt-4 border-t border-[var(--line)] pt-4">
      <summary className="cursor-pointer text-sm font-semibold text-[var(--danger)]">Close fixed deposit</summary>
      <p className="mt-3 text-xs leading-5 text-[var(--muted)]">Closure is permanent. It records a sell event that removes the deposit unit and credits the actual proceeds.</p>
      <form action={action} className="mt-4 grid gap-3" noValidate>
        <input name="portfolioId" type="hidden" value={portfolioID} />
        <input name="accountId" type="hidden" value={accountID} />
        <input name="fixedDepositId" type="hidden" value={fixedDepositID} />
        <Field label="Closure type" error={state.fields?.closureType}>
          <select className="planning-input" defaultValue="" name="closureType">
            <option disabled value="">Select type</option>
            <option value="maturity">Maturity</option>
            <option value="premature">Premature closure</option>
          </select>
        </Field>
        <Field label="Closure date" error={state.fields?.closedAt}><input className="planning-input" name="closedAt" type="date" /></Field>
        <Field label={`Actual proceeds (${currency})`} error={state.fields?.proceeds}><input className="planning-input" min="0" name="proceeds" step="0.01" type="number" /></Field>
        <Field label="Note (optional)"><input className="planning-input" name="note" placeholder="Bank receipt or penalty details" /></Field>
        {state.message && <p aria-live="polite" className={`rounded-xl px-3 py-2 text-xs ${state.success ? "bg-[#edf7f1] text-[var(--brand)]" : "bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>}
        <button className="focus-ring justify-self-start rounded-xl bg-[var(--danger)] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Closing…" : "Record closure"}</button>
      </form>
    </details>
  );
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>;
}
