"use client";

import { useActionState, useEffect, useRef, useState } from "react";
import { createTransactionAction } from "@/app/actions/transactions";
import type { Asset, FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreateTransactionForm({
  portfolioID,
  accountID,
  currency,
  assets,
}: {
  portfolioID: string;
  accountID: string;
  currency: string;
  assets: Asset[];
}) {
  const [state, action, pending] = useActionState(createTransactionAction, initialState);
  const [transactionType, setTransactionType] = useState("deposit");
  const formRef = useRef<HTMLFormElement>(null);
  const occurredAtRef = useRef<HTMLInputElement>(null);
  const occurredAtUTCRef = useRef<HTMLInputElement>(null);
  const usesAsset = transactionType === "buy" || transactionType === "sell";

  useEffect(() => {
    if (state.success) formRef.current?.reset();
  }, [state.success]);

  return (
    <form
      ref={formRef}
      action={action}
      className="mt-6 space-y-4"
      noValidate
      onSubmit={() => {
        if (occurredAtRef.current?.value && occurredAtUTCRef.current) {
          occurredAtUTCRef.current.value = new Date(
            occurredAtRef.current.value,
          ).toISOString();
        }
      }}
    >
      <input name="portfolioId" type="hidden" value={portfolioID} />
      <input name="accountId" type="hidden" value={accountID} />
      <input name="currency" type="hidden" value={currency} />
      <input ref={occurredAtUTCRef} name="occurredAt" type="hidden" />

      <Field label="Event type" error={state.fields?.transactionType}>
        <select
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none"
          name="transactionType"
          value={transactionType}
          onChange={(event) => setTransactionType(event.target.value)}
        >
          <option value="deposit">Deposit</option>
          <option value="withdrawal">Withdrawal</option>
          <option value="buy">Asset purchase</option>
          <option value="sell">Asset sale</option>
          <option value="fee">Fee</option>
          <option value="tax">Tax</option>
        </select>
      </Field>

      <Field label="Date and time" error={state.fields?.occurredAt}>
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none"
          name="occurredAtLocal"
          ref={occurredAtRef}
          type="datetime-local"
        />
      </Field>

      {usesAsset && (
        <>
          <Field label="Asset" error={state.fields?.assetId}>
            <select
              className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none"
              defaultValue=""
              name="assetId"
            >
              <option disabled value="">Choose an asset</option>
              {assets.map((asset) => (
                <option key={asset.id} value={asset.id}>
                  {asset.symbol} · {asset.name}
                </option>
              ))}
            </select>
          </Field>
          <Field label="Quantity" error={state.fields?.quantity}>
            <input
              className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none"
              inputMode="decimal"
              min="0"
              name="quantity"
              placeholder="10"
              step="any"
              type="number"
            />
          </Field>
        </>
      )}

      <Field
        label={usesAsset ? `Total cash amount (${currency})` : `Amount (${currency})`}
        error={state.fields?.amount}
      >
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none"
          inputMode="decimal"
          min="0"
          name="amount"
          placeholder="10000.00"
          step="0.01"
          type="number"
        />
      </Field>

      <Field label="Description (optional)">
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none"
          name="description"
          placeholder="Monthly contribution"
        />
      </Field>

      {usesAsset && assets.length === 0 && (
        <p className="rounded-xl border border-[#e5d3ae] bg-[#fff8e8] px-4 py-3 text-sm text-[#76551f]">
          No active {currency} assets are available. Add asset reference data before recording a purchase or sale.
        </p>
      )}
      {state.message && (
        <p
          className={`rounded-xl border px-4 py-3 text-sm ${state.success ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]" : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"}`}
          role="status"
        >
          {state.message}
        </p>
      )}
      <button
        className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white transition hover:bg-[var(--brand-strong)] disabled:cursor-wait disabled:opacity-65"
        disabled={pending || (usesAsset && assets.length === 0)}
        type="submit"
      >
        {pending ? "Recording…" : "Record transaction"}
      </button>
    </form>
  );
}

function Field({ label, error, children }: { label: string; error?: string; children: React.ReactNode }) {
  return (
    <div>
      <label className="mb-2 block text-sm font-semibold">{label}</label>
      {children}
      {error && <p className="mt-1.5 text-xs text-[var(--danger)]">{error}</p>}
    </div>
  );
}
