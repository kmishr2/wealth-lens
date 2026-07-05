"use client";

import { useActionState } from "react";
import { deletePortfolioAction, updatePortfolioAction } from "@/app/actions/portfolios";
import type { FormState, Portfolio } from "@/lib/types";

const initialState: FormState = { message: "" };

export function PortfolioSettings({ portfolio }: { portfolio: Portfolio }) {
  const [updateState, updateAction, updating] = useActionState(updatePortfolioAction, initialState);
  const [deleteState, deleteAction, deleting] = useActionState(deletePortfolioAction, initialState);
  return <div className="mt-8 grid gap-4 lg:grid-cols-2">
    <details className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] p-5">
      <summary className="cursor-pointer font-semibold text-[var(--brand)]">Edit portfolio details</summary>
      <form action={updateAction} className="mt-4 space-y-4">
        <input name="portfolioId" type="hidden" value={portfolio.id} />
        <label className="text-sm font-semibold">Name<input className="planning-input" defaultValue={portfolio.name} name="name" /></label>
        <label className="text-sm font-semibold">Description<textarea className="planning-input min-h-24" defaultValue={portfolio.description} name="description" /></label>
        {updateState.message && <StateMessage state={updateState} />}
        <button className="focus-ring rounded-xl bg-[var(--brand)] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60" disabled={updating} type="submit">{updating ? "Saving…" : "Save changes"}</button>
      </form>
    </details>
    <details className="rounded-2xl border border-[#e8c9c4] bg-[#fff9f7] p-5">
      <summary className="cursor-pointer font-semibold text-[var(--danger)]">Delete portfolio</summary>
      <form action={deleteAction} className="mt-4 space-y-4">
        <input name="portfolioId" type="hidden" value={portfolio.id} /><input name="expectedName" type="hidden" value={portfolio.name} />
        <p className="text-sm leading-6 text-[var(--muted)]">This removes the portfolio from active use. Type <strong>{portfolio.name}</strong> exactly to continue.</p>
        <label className="text-sm font-semibold">Confirmation<input autoComplete="off" className="planning-input" name="confirmation" /></label>
        {deleteState.message && <StateMessage state={deleteState} />}
        <button className="focus-ring rounded-xl border border-[#d9aaa3] px-4 py-2.5 text-sm font-semibold text-[var(--danger)] disabled:opacity-60" disabled={deleting} type="submit">{deleting ? "Deleting…" : "Delete portfolio"}</button>
      </form>
    </details>
  </div>;
}

function StateMessage({ state }: { state: FormState }) { return <p className={`rounded-xl px-4 py-3 text-sm ${state.success ? "bg-[#edf7f1] text-[var(--brand)]" : "bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>; }
