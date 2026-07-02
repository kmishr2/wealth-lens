"use client";

import { useActionState, useEffect, useRef } from "react";
import { createAssetAction } from "@/app/actions/assets";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreateAssetForm({ defaultCurrency }: { defaultCurrency: string }) {
  const [state, action, pending] = useActionState(createAssetAction, initialState);
  const formRef = useRef<HTMLFormElement>(null);
  useEffect(() => { if (state.success) formRef.current?.reset(); }, [state.success]);
  return (
    <form ref={formRef} action={action} className="mt-5 space-y-4" noValidate>
      <AssetField label="Symbol" error={state.fields?.symbol}><input className="planning-input" name="symbol" placeholder="NIFTYBEES" /></AssetField>
      <AssetField label="Name" error={state.fields?.name}><input className="planning-input" name="name" placeholder="Nippon India ETF Nifty 50 BeES" /></AssetField>
      <div className="grid gap-4 sm:grid-cols-2">
        <AssetField label="Asset class" error={state.fields?.assetClass}>
          <select className="planning-input" defaultValue="equity" name="assetClass">
            <option value="equity">Equity</option><option value="fund">Fund</option><option value="bond">Bond</option><option value="cash">Cash</option><option value="commodity">Commodity</option><option value="real_estate">Real estate</option><option value="crypto">Crypto</option><option value="alternative">Alternative</option><option value="other">Other</option>
          </select>
        </AssetField>
        <AssetField label="Risk category" error={state.fields?.riskCategory}>
          <select className="planning-input" defaultValue="" name="riskCategory">
            <option value="">Automatic / unclassified</option><option value="equity">Equity</option><option value="debt">Debt</option><option value="cash_other">Cash / other</option>
          </select>
        </AssetField>
      </div>
      <div className="grid gap-4 sm:grid-cols-2">
        <AssetField label="Currency" error={state.fields?.currency}><input className="planning-input uppercase" defaultValue={defaultCurrency} maxLength={3} name="currency" /></AssetField>
        <AssetField label="Exchange"><input className="planning-input uppercase" name="exchange" placeholder="NSE" /></AssetField>
      </div>
      {state.message && <p className={`rounded-xl border px-4 py-3 text-sm ${state.success ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]" : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>}
      <button className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Creating…" : "Create asset"}</button>
    </form>
  );
}

function AssetField({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>;
}
