"use client";

import { useActionState } from "react";
import { deleteGoalAction, updateGoalAction } from "@/app/actions/goals";
import type { FormState, Goal } from "@/lib/types";

const initialState: FormState = { message: "" };

export function GoalActions({
  goal,
  portfolioID,
}: {
  goal: Goal;
  portfolioID: string;
}) {
  const [updateState, updateAction, updating] = useActionState(
    updateGoalAction,
    initialState,
  );
  const [deleteState, deleteAction, deleting] = useActionState(
    deleteGoalAction,
    initialState,
  );

  return (
    <details className="mt-5 border-t border-[var(--line)] pt-4">
      <summary className="cursor-pointer text-sm font-semibold text-[var(--brand)]">
        Manage goal
      </summary>
      <form action={updateAction} className="mt-4 grid gap-3" noValidate>
        <GoalIDs goalID={goal.id} portfolioID={portfolioID} />
        <Field label="Name" error={updateState.fields?.name}>
          <input className="planning-input" defaultValue={goal.name} name="name" />
        </Field>
        <div className="grid gap-3 sm:grid-cols-2">
          <Field label="Target amount" error={updateState.fields?.targetAmount}>
            <input className="planning-input" defaultValue={goal.target_amount} min="0" name="targetAmount" step="0.01" type="number" />
          </Field>
          <Field label="Currency" error={updateState.fields?.currency}>
            <input className="planning-input uppercase" defaultValue={goal.currency} maxLength={3} name="currency" />
          </Field>
          <Field label="Target date" error={updateState.fields?.targetDate}>
            <input className="planning-input" defaultValue={goal.target_date} name="targetDate" type="date" />
          </Field>
          <Field label="Status" error={updateState.fields?.status}>
            <select className="planning-input" defaultValue={goal.status} name="status">
              <option value="active">Active</option>
              <option value="completed">Completed</option>
              <option value="archived">Archived</option>
            </select>
          </Field>
        </div>
        <p className="text-xs leading-5 text-[var(--muted)]">
          Changes apply to future progress snapshots. Existing snapshots remain unchanged.
        </p>
        {updateState.message && <StateMessage state={updateState} />}
        <button className="focus-ring justify-self-start rounded-xl bg-[var(--brand)] px-4 py-2.5 text-sm font-semibold text-white disabled:opacity-60" disabled={updating} type="submit">
          {updating ? "Saving…" : "Save goal"}
        </button>
      </form>

      <form action={deleteAction} className="mt-5 border-t border-[#e8c9c4] pt-4">
        <GoalIDs goalID={goal.id} portfolioID={portfolioID} />
        <input name="expectedName" type="hidden" value={goal.name} />
        <Field label={`Type “${goal.name}” to delete`} error={deleteState.fields?.confirmation}>
          <input autoComplete="off" className="planning-input" name="confirmation" />
        </Field>
        {deleteState.message && <div className="mt-3"><StateMessage state={deleteState} /></div>}
        <button className="focus-ring mt-3 rounded-xl border border-[#d9aaa3] px-4 py-2.5 text-sm font-semibold text-[var(--danger)] disabled:opacity-60" disabled={deleting} type="submit">
          {deleting ? "Deleting…" : "Delete goal"}
        </button>
      </form>
    </details>
  );
}

function GoalIDs({ goalID, portfolioID }: { goalID: string; portfolioID: string }) {
  return <><input name="portfolioId" type="hidden" value={portfolioID} /><input name="goalId" type="hidden" value={goalID} /></>;
}

function Field({ children, error, label }: { children: React.ReactNode; error?: string; label: string }) {
  return <label className="text-sm font-semibold">{label}{children}{error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}</label>;
}

function StateMessage({ state }: { state: FormState }) {
  return <p aria-live="polite" className={`rounded-xl px-4 py-3 text-sm ${state.success ? "bg-[#edf7f1] text-[var(--brand)]" : "bg-[#fff4f2] text-[var(--danger)]"}`}>{state.message}</p>;
}
