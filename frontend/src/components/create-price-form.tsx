"use client";

import { useActionState, useEffect, useRef } from "react";
import { createPriceAction } from "@/app/actions/prices";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreatePriceForm({ assetID, currency }: { assetID: string; currency: string }) {
  const [state, action, pending] = useActionState(createPriceAction, initialState);
  const formRef = useRef<HTMLFormElement>(null);
  const localTimeRef = useRef<HTMLInputElement>(null);
  const utcTimeRef = useRef<HTMLInputElement>(null);
  useEffect(() => { if (state.success) formRef.current?.reset(); }, [state.success]);
  return (
    <form
      ref={formRef}
      action={action}
      className="mt-5 space-y-4"
      noValidate
      onSubmit={() => {
        if (localTimeRef.current?.value && utcTimeRef.current) utcTimeRef.current.value = new Date(localTimeRef.current.value).toISOString();
      }}
    >
      <input name="assetId" type="hidden" value={assetID} /><input name="currency" type="hidden" value={currency} /><input ref={utcTimeRef} name="pricedAt" type="hidden" />
      <PriceField label={`Price (${currency})`} error={state.fields?.price}><input className="planning-input" min="0" name="price" step="0.0001" type="number" /></PriceField>
      <PriceField label="Price date and time" error={state.fields?.pricedAt}><input ref={localTimeRef} className="planning-input" name="pricedAtLocal" type="datetime-local" /></PriceField>
      <PriceField label="Source"><input className="planning-input" defaultValue="manual" name="source" /></PriceField>
      <PriceField label="Note (optional)"><input className="planning-input" name="note" /></PriceField>
      {state.message && <p className={`rounded-xl border px-4 py-3 text-sm ${state.success ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]" : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>}
      <button className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Recording…" : "Record price"}</button>
    </form>
  );
}

function PriceField({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>;
}
