"use client";

import { useActionState, useEffect, useRef } from "react";
import { createPortfolioAction } from "@/app/actions/portfolios";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreatePortfolioForm({
  defaultCurrency,
}: {
  defaultCurrency: string;
}) {
  const [state, action, pending] = useActionState(
    createPortfolioAction,
    initialState,
  );
  const formRef = useRef<HTMLFormElement>(null);

  useEffect(() => {
    if (state.success) formRef.current?.reset();
  }, [state.success]);

  return (
    <form ref={formRef} action={action} className="mt-6 space-y-4" noValidate>
      <div>
        <label className="mb-2 block text-sm font-semibold" htmlFor="name">
          Portfolio name
        </label>
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none transition focus:border-[var(--brand)]"
          id="name"
          name="name"
          placeholder="Long-term portfolio"
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
          htmlFor="description"
        >
          Description <span className="font-normal text-[var(--muted)]">(optional)</span>
        </label>
        <textarea
          className="focus-ring min-h-24 w-full resize-y rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none transition focus:border-[var(--brand)]"
          id="description"
          name="description"
          placeholder="What this portfolio is for"
        />
      </div>
      <div>
        <label
          className="mb-2 block text-sm font-semibold"
          htmlFor="baseCurrency"
        >
          Base currency
        </label>
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 uppercase outline-none transition focus:border-[var(--brand)]"
          id="baseCurrency"
          name="baseCurrency"
          defaultValue={defaultCurrency}
          maxLength={3}
        />
        {state.fields?.baseCurrency && (
          <p className="mt-1.5 text-xs text-[var(--danger)]">
            {state.fields.baseCurrency}
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
        type="submit"
        disabled={pending}
      >
        {pending ? "Creating…" : "Create portfolio"}
      </button>
    </form>
  );
}
