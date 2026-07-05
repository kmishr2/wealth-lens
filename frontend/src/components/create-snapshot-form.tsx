"use client";

import { useActionState } from "react";
import { createDailySnapshotAction } from "@/app/actions/snapshots";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function CreateSnapshotForm({ portfolioID }: { portfolioID: string }) {
  const [state, action, pending] = useActionState(
    createDailySnapshotAction,
    initialState,
  );

  return (
    <form
      action={action}
      className="mt-5 grid gap-4 sm:grid-cols-[minmax(0,240px)_auto] sm:items-end"
      noValidate
    >
      <input name="portfolioId" type="hidden" value={portfolioID} />
      <label className="text-sm font-semibold">
        Snapshot date
        <input className="planning-input" name="snapshotDate" type="date" />
        {state.fields?.snapshotDate && (
          <span className="mt-1 block text-xs text-[var(--danger)]">
            {state.fields.snapshotDate}
          </span>
        )}
      </label>
      <button
        className="focus-ring rounded-xl bg-[var(--brand)] px-5 py-3 font-semibold text-white disabled:opacity-60"
        disabled={pending}
        type="submit"
      >
        {pending ? "Creating…" : "Create daily snapshot"}
      </button>
      {state.message && (
        <p
          aria-live="polite"
          className={`rounded-xl border px-4 py-3 text-sm sm:col-span-2 ${
            state.success
              ? "border-[#b9d6c5] bg-[#edf7f1] text-[var(--brand)]"
              : "border-[#e8c9c4] bg-[#fff4f2] text-[var(--danger)]"
          }`}
        >
          {state.message}
        </p>
      )}
    </form>
  );
}
