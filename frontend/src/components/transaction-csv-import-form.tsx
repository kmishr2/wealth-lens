"use client";

import { useActionState, useEffect, useRef } from "react";
import { importTransactionsCSVAction } from "@/app/actions/transactions";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };
const header = "transaction_type,occurred_at,description,asset_id,quantity,amount,currency,idempotency_key";

export function TransactionCSVImportForm({ portfolioID, accountID }: { portfolioID: string; accountID: string }) {
  const [state, action, pending] = useActionState(importTransactionsCSVAction, initialState);
  const formRef = useRef<HTMLFormElement>(null);
  useEffect(() => { if (state.success) formRef.current?.reset(); }, [state.success]);
  return (
    <form action={action} className="mt-5 space-y-4" encType="multipart/form-data" ref={formRef}>
      <input name="portfolioId" type="hidden" value={portfolioID} />
      <input name="accountId" type="hidden" value={accountID} />
      <label className="block text-sm font-semibold">
        CSV file
        <input accept=".csv,text/csv" className="planning-input" name="file" required type="file" />
        {state.fields?.file && <span className="mt-1 block text-xs text-[var(--danger)]">{state.fields.file}</span>}
      </label>
      <div className="overflow-x-auto rounded-xl bg-[#f1f2ed] p-3">
        <code className="whitespace-nowrap text-xs text-[var(--muted)]">{header}</code>
      </div>
      <p className="text-xs leading-5 text-[var(--muted)]">Up to 1,000 rows. Timestamps use RFC3339. Amounts and quantities are positive; the backend applies ledger signs and rejects the entire file if any row fails.</p>
      {state.message && <p aria-live="polite" className={`rounded-xl border px-4 py-3 text-sm ${state.success ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]" : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>}
      <button className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white disabled:opacity-60" disabled={pending} type="submit">{pending ? "Importing…" : "Import CSV"}</button>
    </form>
  );
}
