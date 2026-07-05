"use client";

import { useActionState, useEffect, useRef } from "react";
import { createBenchmarkObservationAction } from "@/app/actions/benchmarks";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreateBenchmarkObservationForm({ benchmarkID }: { benchmarkID: string }) {
  const [state, action, pending] = useActionState(createBenchmarkObservationAction, initialState);
  const formRef = useRef<HTMLFormElement>(null);
  useEffect(() => { if (state.success) formRef.current?.reset(); }, [state.success]);
  return <form ref={formRef} action={action} className="mt-5 space-y-4" noValidate>
    <input name="benchmarkId" type="hidden" value={benchmarkID} />
    <Field label="Observation date" error={state.fields?.observationDate}><input className="planning-input" name="observationDate" type="date" /></Field>
    <Field label="Index value" error={state.fields?.value}><input className="planning-input" min="0" name="value" step="0.0001" type="number" /></Field>
    <Field label="Source"><input className="planning-input" defaultValue="manual" name="source" /></Field>
    <Field label="Note (optional)"><input className="planning-input" name="note" /></Field>
    {state.message && <p className={`rounded-xl border px-4 py-3 text-sm ${state.success ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]" : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>}
    <button className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Recording…" : "Record observation"}</button>
  </form>;
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) { return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>; }
