"use client";

import { useActionState } from "react";
import { deleteAccountAction, updateAccountAction } from "@/app/actions/accounts";
import type { Account, FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

export function AccountSettings({ account }: { account: Account }) {
  const [updateState, updateAction, updating] = useActionState(updateAccountAction, initialState);
  const [deleteState, deleteAction, deleting] = useActionState(deleteAccountAction, initialState);
  return <div className="mt-8 grid gap-4 lg:grid-cols-2">
    <details className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] p-5">
      <summary className="cursor-pointer font-semibold text-[var(--brand)]">Edit account details</summary>
      <form action={updateAction} className="mt-4 space-y-4">
        <IDs account={account} />
        <label className="text-sm font-semibold">Name<input className="planning-input" defaultValue={account.name} name="name" /></label>
        <label className="text-sm font-semibold">Institution<input className="planning-input" defaultValue={account.institution_name} name="institutionName" /></label>
        {updateState.message && <StateMessage state={updateState} />}
        <button className="focus-ring rounded-xl bg-[var(--brand)] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60" disabled={updating} type="submit">{updating ? "Saving…" : "Save changes"}</button>
      </form>
    </details>
    <details className="rounded-2xl border border-[#e8c9c4] bg-[#fff9f7] p-5">
      <summary className="cursor-pointer font-semibold text-[var(--danger)]">Delete account</summary>
      <form action={deleteAction} className="mt-4 space-y-4">
        <IDs account={account} /><input name="expectedName" type="hidden" value={account.name} />
        <p className="text-sm leading-6 text-[var(--muted)]">Type <strong>{account.name}</strong> exactly. Existing ledger relationships may prevent deletion.</p>
        <label className="text-sm font-semibold">Confirmation<input autoComplete="off" className="planning-input" name="confirmation" /></label>
        {deleteState.message && <StateMessage state={deleteState} />}
        <button className="focus-ring rounded-xl border border-[#d9aaa3] px-4 py-2.5 text-sm font-semibold text-[var(--danger)] disabled:opacity-60" disabled={deleting} type="submit">{deleting ? "Deleting…" : "Delete account"}</button>
      </form>
    </details>
  </div>;
}

function IDs({ account }: { account: Account }) { return <><input name="portfolioId" type="hidden" value={account.portfolio_id} /><input name="accountId" type="hidden" value={account.id} /></>; }
function StateMessage({ state }: { state: FormState }) { return <p className={`rounded-xl px-4 py-3 text-sm ${state.success ? "bg-[#edf7f1] text-[var(--brand)]" : "bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>; }
