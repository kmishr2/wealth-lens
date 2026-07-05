"use client";

import { useActionState, useEffect, useRef } from "react";
import { createFixedDepositAction } from "@/app/actions/fixed-deposits";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreateFixedDepositForm({ portfolioID, accountID, currency }: { portfolioID: string; accountID: string; currency: string }) {
  const [state, action, pending] = useActionState(createFixedDepositAction, initialState);
  const formRef = useRef<HTMLFormElement>(null);
  useEffect(() => { if (state.success) formRef.current?.reset(); }, [state.success]);

  return (
    <form action={action} className="mt-5 grid gap-4 sm:grid-cols-2" noValidate ref={formRef}>
      <input name="portfolioId" type="hidden" value={portfolioID} />
      <input name="accountId" type="hidden" value={accountID} />
      <input name="currency" type="hidden" value={currency} />
      <Field label="Deposit name" error={state.fields?.name}><input className="planning-input" name="name" placeholder="12-month fixed deposit" /></Field>
      <Field label="Bank reference (optional)"><input className="planning-input" name="bankReference" /></Field>
      <Field label={`Principal (${currency})`} error={state.fields?.principal}><input className="planning-input" min="0" name="principal" step="0.01" type="number" /></Field>
      <Field label="Annual ROI (%)" error={state.fields?.annualInterestRate}><input className="planning-input" max="100" min="0" name="annualInterestRate" step="0.0001" type="number" /></Field>
      <Field label="Start date" error={state.fields?.startDate}><input className="planning-input" name="startDate" type="date" /></Field>
      <Field label="Maturity date" error={state.fields?.maturityDate}><input className="planning-input" name="maturityDate" type="date" /></Field>
      <Field label={`Current value (${currency})`} error={state.fields?.currentValue}><input className="planning-input" min="0" name="currentValue" step="0.01" type="number" /></Field>
      <Field label="Current value date" error={state.fields?.currentValueDate}><input className="planning-input" name="currentValueDate" type="date" /></Field>
      <p className="text-xs leading-5 text-[var(--muted)] sm:col-span-2">ROI is stored as contract metadata. Current value is explicit and is not estimated from an assumed compounding schedule.</p>
      {state.message && <p aria-live="polite" className={`rounded-xl border px-4 py-3 text-sm sm:col-span-2 ${state.success ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]" : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>}
      <button className="focus-ring rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white disabled:opacity-60 sm:col-span-2" disabled={pending} type="submit">{pending ? "Adding…" : "Add fixed deposit"}</button>
    </form>
  );
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>;
}
