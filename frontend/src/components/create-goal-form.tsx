"use client";

import { useActionState, useEffect, useRef } from "react";
import { createGoalAction } from "@/app/actions/goals";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreateGoalForm({ portfolioID, currency }: { portfolioID: string; currency: string }) {
  const [state, action, pending] = useActionState(createGoalAction, initialState);
  const formRef = useRef<HTMLFormElement>(null);
  useEffect(() => { if (state.success) formRef.current?.reset(); }, [state.success]);
  return (
    <form ref={formRef} action={action} className="mt-5 grid gap-4 sm:grid-cols-2" noValidate>
      <input name="portfolioId" type="hidden" value={portfolioID} />
      <GoalField label="Goal name" error={state.fields?.name}><input className="planning-input" name="name" placeholder="Home down payment" /></GoalField>
      <GoalField label={`Target amount (${currency})`} error={state.fields?.targetAmount}><input className="planning-input" min="0" name="targetAmount" step="0.01" type="number" /></GoalField>
      <input name="currency" type="hidden" value={currency} />
      <GoalField label="Target date" error={state.fields?.targetDate}><input className="planning-input" name="targetDate" type="date" /></GoalField>
      <div className="self-end"><button className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Creating…" : "Create goal"}</button></div>
      {state.message && <p className={`sm:col-span-2 rounded-xl border px-4 py-3 text-sm ${state.success ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]" : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>}
    </form>
  );
}

function GoalField({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>;
}
