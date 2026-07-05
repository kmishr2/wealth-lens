"use client";

import { useActionState, useRef, useState } from "react";
import { correctTransactionAction, reverseTransactionAction } from "@/app/actions/transactions";
import type { Asset, FormState, Transaction } from "@/lib/types";

const initialState: FormState = { message: "" };

export function TransactionAuditActions({
  portfolioID,
  accountID,
  currency,
  transaction,
  assets,
  superseded,
}: {
  portfolioID: string;
  accountID: string;
  currency: string;
  transaction: Transaction;
  assets: Asset[];
  superseded: boolean;
}) {
  const [reverseState, reverseAction, reversing] = useActionState(reverseTransactionAction, initialState);
  const [correctState, correctAction, correcting] = useActionState(correctTransactionAction, initialState);
  const [type, setType] = useState<string>(
    transaction.transaction_type === "transfer"
      ? "deposit"
      : transaction.transaction_type,
  );
  const localTimeRef = useRef<HTMLInputElement>(null);
  const utcTimeRef = useRef<HTMLInputElement>(null);
  const usesAsset = type === "buy" || type === "sell";
  const unavailable =
    superseded ||
    transaction.transaction_type === "reversal" ||
    transaction.transaction_type === "transfer";

  if (unavailable) {
    const label = transaction.transaction_type === "reversal"
      ? "Audit reversal event"
      : transaction.transaction_type === "transfer"
        ? "Transfer correction requires the multi-account workflow"
        : "Original event superseded";
    return <p className="mt-4 border-t border-[var(--line)] pt-4 text-xs font-semibold text-[var(--muted)]">{label}</p>;
  }

  return (
    <div className="mt-4 grid gap-3 border-t border-[var(--line)] pt-4 sm:grid-cols-2">
      <details className="rounded-xl border border-[var(--line)] bg-[var(--surface)] p-3">
        <summary className="cursor-pointer text-sm font-semibold text-[var(--danger)]">Reverse event</summary>
        <form action={reverseAction} className="mt-3 space-y-3">
          <AuditIDs portfolioID={portfolioID} accountID={accountID} transactionID={transaction.id} />
          <label className="text-xs font-semibold">Reason<input className="planning-input" name="reason" placeholder="Duplicate entry" /></label>
          {reverseState.message && <AuditMessage state={reverseState} />}
          <button className="focus-ring rounded-lg border border-[#d9aaa3] px-3 py-2 text-xs font-bold text-[var(--danger)] disabled:opacity-60" disabled={reversing} type="submit">{reversing ? "Reversing…" : "Create reversal"}</button>
        </form>
      </details>

      <details className="rounded-xl border border-[var(--line)] bg-[var(--surface)] p-3">
        <summary className="cursor-pointer text-sm font-semibold text-[var(--brand)]">Correct event</summary>
        <form
          action={correctAction}
          className="mt-3 space-y-3"
          onSubmit={() => {
            if (localTimeRef.current?.value && utcTimeRef.current) utcTimeRef.current.value = new Date(localTimeRef.current.value).toISOString();
          }}
        >
          <AuditIDs portfolioID={portfolioID} accountID={accountID} transactionID={transaction.id} />
          <input name="currency" type="hidden" value={currency} /><input ref={utcTimeRef} name="occurredAt" type="hidden" />
          <label className="text-xs font-semibold">Correction reason<input className="planning-input" name="reason" placeholder="Incorrect amount" /></label>
          <label className="text-xs font-semibold">Replacement type<select className="planning-input" name="transactionType" value={type} onChange={(event) => setType(event.target.value)}><option value="deposit">Deposit</option><option value="withdrawal">Withdrawal</option><option value="buy">Asset purchase</option><option value="sell">Asset sale</option><option value="fee">Fee</option><option value="tax">Tax</option></select></label>
          <label className="text-xs font-semibold">Replacement date<input ref={localTimeRef} className="planning-input" name="occurredAtLocal" type="datetime-local" /></label>
          {usesAsset && <>
            <label className="text-xs font-semibold">Asset<select className="planning-input" defaultValue="" name="assetId"><option disabled value="">Choose an asset</option>{assets.map((asset) => <option key={asset.id} value={asset.id}>{asset.symbol} · {asset.name}</option>)}</select></label>
            <label className="text-xs font-semibold">Quantity<input className="planning-input" min="0" name="quantity" step="any" type="number" /></label>
          </>}
          <label className="text-xs font-semibold">{usesAsset ? "Total cash amount" : "Amount"}<input className="planning-input" min="0" name="amount" step="0.01" type="number" /></label>
          <label className="text-xs font-semibold">Description<input className="planning-input" name="description" /></label>
          {correctState.message && <AuditMessage state={correctState} />}
          <button className="focus-ring rounded-lg bg-[var(--brand)] px-3 py-2 text-xs font-bold text-white disabled:opacity-60" disabled={correcting || (usesAsset && assets.length === 0)} type="submit">{correcting ? "Correcting…" : "Reverse and replace"}</button>
        </form>
      </details>
    </div>
  );
}

function AuditIDs({ portfolioID, accountID, transactionID }: { portfolioID: string; accountID: string; transactionID: string }) {
  return <><input name="portfolioId" type="hidden" value={portfolioID} /><input name="accountId" type="hidden" value={accountID} /><input name="transactionId" type="hidden" value={transactionID} /></>;
}

function AuditMessage({ state }: { state: FormState }) {
  return <p className={`rounded-lg px-3 py-2 text-xs ${state.success ? "bg-[#edf7f1] text-[var(--brand)]" : "bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>;
}
