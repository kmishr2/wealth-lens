"use client";

import { useActionState } from "react";
import { updateProfileAction } from "@/app/actions/users";
import type { FormState, User } from "@/lib/types";

const initialState: FormState = { message: "" };

export function ProfileForm({ user }: { user: User }) {
  const [state, action, pending] = useActionState(
    updateProfileAction,
    initialState,
  );

  return (
    <form action={action} className="mt-7 space-y-5">
      <Field label="Display name" error={state.fields?.displayName}>
        <input
          className="planning-input"
          defaultValue={user.display_name}
          name="displayName"
          required
        />
      </Field>

      <Field label="Base currency" error={state.fields?.baseCurrency}>
        <input
          autoCapitalize="characters"
          className="planning-input uppercase"
          defaultValue={user.base_currency}
          maxLength={3}
          minLength={3}
          name="baseCurrency"
          pattern="[A-Za-z]{3}"
          required
        />
      </Field>

      <Field label="Timezone" error={state.fields?.timezone}>
        <input
          className="planning-input"
          defaultValue={user.timezone}
          name="timezone"
          placeholder="Asia/Kolkata"
          required
        />
      </Field>

      {state.message && (
        <p
          aria-live="polite"
          className={`rounded-xl px-4 py-3 text-sm ${
            state.success
              ? "bg-[#edf7f1] text-[var(--brand)]"
              : "bg-[#fff4f2] text-[var(--danger)]"
          }`}
        >
          {state.message}
        </p>
      )}

      <button
        className="focus-ring rounded-xl bg-[var(--brand)] px-5 py-3 font-semibold text-white disabled:opacity-60"
        disabled={pending}
        type="submit"
      >
        {pending ? "Saving…" : "Save profile"}
      </button>
    </form>
  );
}

function Field({
  children,
  error,
  label,
}: {
  children: React.ReactNode;
  error?: string;
  label: string;
}) {
  return (
    <label className="block text-sm font-semibold">
      {label}
      {children}
      {error && <span className="mt-1 block text-xs text-[var(--danger)]">{error}</span>}
    </label>
  );
}
