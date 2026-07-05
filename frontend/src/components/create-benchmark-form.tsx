"use client";

import { useActionState, useEffect, useRef } from "react";
import { createBenchmarkAction } from "@/app/actions/benchmarks";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreateBenchmarkForm({ defaultCurrency }: { defaultCurrency: string }) {
  const [state, action, pending] = useActionState(createBenchmarkAction, initialState);
  const formRef = useRef<HTMLFormElement>(null);
  useEffect(() => { if (state.success) formRef.current?.reset(); }, [state.success]);
  return <form ref={formRef} action={action} className="mt-5 space-y-4" noValidate>
    <Field label="Code" error={state.fields?.code}><input className="planning-input uppercase" name="code" placeholder="NIFTY50" /></Field>
    <Field label="Name" error={state.fields?.name}><input className="planning-input" name="name" placeholder="Nifty 50" /></Field>
    <div className="grid gap-4 sm:grid-cols-2"><Field label="Currency" error={state.fields?.currency}><input className="planning-input uppercase" defaultValue={defaultCurrency} maxLength={3} name="currency" /></Field><Field label="Source" error={state.fields?.source}><input className="planning-input" name="source" placeholder="NSE manual export" /></Field></div>
    <Field label="Description"><textarea className="planning-input min-h-24" name="description" /></Field>
    {state.message && <Message state={state} />}
    <button className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Creating…" : "Create benchmark"}</button>
  </form>;
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) { return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>; }
function Message({ state }: { state: FormState }) { return <p className={`rounded-xl border px-4 py-3 text-sm ${state.success ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]" : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>; }
