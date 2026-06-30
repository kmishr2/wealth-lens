"use client";

import { useActionState, useEffect, useRef } from "react";
import { createAccountAction } from "@/app/actions/accounts";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreateAccountForm({
  portfolioID,
  defaultCurrency,
}: {
  portfolioID: string;
  defaultCurrency: string;
}) {
  const [state, action, pending] = useActionState(
    createAccountAction,
    initialState,
  );
  const formRef = useRef<HTMLFormElement>(null);

  useEffect(() => {
    if (state.success) formRef.current?.reset();
  }, [state.success]);

  return (
    <form ref={formRef} action={action} className="mt-6 space-y-4" noValidate>
      <input name="portfolioId" type="hidden" value={portfolioID} />

      <div>
        <label className="mb-2 block text-sm font-semibold" htmlFor="name">
          Account name
        </label>
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none transition focus:border-[var(--brand)]"
          id="name"
          name="name"
          placeholder="Primary brokerage"
        />
        {state.fields?.name && (
          <p className="mt-1.5 text-xs text-[var(--danger)]">
            {state.fields.name}
          </p>
        )}
      </div>

      <div>
        <label
          className="mb-2 block text-sm font-semibold"
          htmlFor="accountType"
        >
          Account type
        </label>
        <select
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none transition focus:border-[var(--brand)]"
          defaultValue="brokerage"
          id="accountType"
          name="accountType"
        >
          <option value="brokerage">Brokerage</option>
          <option value="retirement">Retirement</option>
          <option value="bank">Bank</option>
          <option value="wallet">Wallet</option>
          <option value="other">Other</option>
        </select>
        {state.fields?.accountType && (
          <p className="mt-1.5 text-xs text-[var(--danger)]">
            {state.fields.accountType}
          </p>
        )}
      </div>

      <div>
        <label
          className="mb-2 block text-sm font-semibold"
          htmlFor="institutionName"
        >
          Institution <span className="font-normal text-[var(--muted)]">(optional)</span>
        </label>
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none transition focus:border-[var(--brand)]"
          id="institutionName"
          name="institutionName"
          placeholder="Broker or bank name"
        />
      </div>

      <div>
        <label className="mb-2 block text-sm font-semibold" htmlFor="currency">
          Currency
        </label>
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 uppercase outline-none transition focus:border-[var(--brand)]"
          defaultValue={defaultCurrency}
          id="currency"
          maxLength={3}
          name="currency"
        />
        {state.fields?.currency && (
          <p className="mt-1.5 text-xs text-[var(--danger)]">
            {state.fields.currency}
          </p>
        )}
      </div>

      {state.message && (
        <p
          className={`rounded-xl border px-4 py-3 text-sm ${
            state.success
              ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]"
              : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"
          }`}
          role="status"
        >
          {state.message}
        </p>
      )}

      <button
        className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3 font-semibold text-white transition hover:bg-[var(--brand-strong)] disabled:cursor-wait disabled:opacity-65"
        disabled={pending}
        type="submit"
      >
        {pending ? "Creating…" : "Create account"}
      </button>
    </form>
  );
}
