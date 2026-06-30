"use client";

import Link from "next/link";
import { useActionState } from "react";
import { loginAction, registerAction } from "@/app/actions/auth";
import type { FormState } from "@/lib/types";

const initialState: FormState = { message: "" };

type AuthFormProps = {
  mode: "login" | "register";
  notice?: string;
};

export function AuthForm({ mode, notice }: AuthFormProps) {
  const action = mode === "login" ? loginAction : registerAction;
  const [state, formAction, pending] = useActionState(action, initialState);
  const isLogin = mode === "login";

  return (
    <form action={formAction} className="mt-8 space-y-5" noValidate>
      {notice && (
        <p className="rounded-xl border border-[#d6d8ce] bg-white px-4 py-3 text-sm text-[var(--muted)]">
          {notice}
        </p>
      )}
      {!isLogin && (
        <Field
          id="displayName"
          label="Display name"
          error={state.fields?.displayName}
        >
          <input
            className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 text-[0.95rem] outline-none transition focus:border-[var(--brand)]"
            id="displayName"
            name="displayName"
            autoComplete="name"
            placeholder="Your name"
          />
        </Field>
      )}

      <Field id="email" label="Email address" error={state.fields?.email}>
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 text-[0.95rem] outline-none transition focus:border-[var(--brand)]"
          id="email"
          name="email"
          type="email"
          autoComplete="email"
          placeholder="you@example.com"
        />
      </Field>

      <Field id="password" label="Password" error={state.fields?.password}>
        <input
          className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 text-[0.95rem] outline-none transition focus:border-[var(--brand)]"
          id="password"
          name="password"
          type="password"
          autoComplete={isLogin ? "current-password" : "new-password"}
          placeholder={isLogin ? "Your password" : "At least 12 characters"}
        />
      </Field>

      {!isLogin && (
        <div className="grid gap-4 sm:grid-cols-2">
          <Field
            id="baseCurrency"
            label="Base currency"
            error={state.fields?.baseCurrency}
          >
            <input
              className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 uppercase outline-none transition focus:border-[var(--brand)]"
              id="baseCurrency"
              name="baseCurrency"
              defaultValue="USD"
              maxLength={3}
            />
          </Field>
          <Field id="timezone" label="Timezone">
            <input
              className="focus-ring w-full rounded-xl border border-[var(--line)] bg-white px-4 py-3 outline-none transition focus:border-[var(--brand)]"
              id="timezone"
              name="timezone"
              defaultValue="UTC"
              placeholder="UTC"
            />
          </Field>
        </div>
      )}

      {state.message && (
        <p
          className="rounded-xl border border-[#e8c9c4] bg-[#fff4f2] px-4 py-3 text-sm text-[var(--danger)]"
          role="alert"
        >
          {state.message}
        </p>
      )}

      <button
        className="focus-ring w-full rounded-xl bg-[var(--brand)] px-4 py-3.5 font-semibold text-white transition hover:bg-[var(--brand-strong)] disabled:cursor-wait disabled:opacity-65"
        type="submit"
        disabled={pending}
      >
        {pending
          ? isLogin
            ? "Signing in…"
            : "Creating account…"
          : isLogin
            ? "Sign in"
            : "Create account"}
      </button>

      <p className="text-center text-sm text-[var(--muted)]">
        {isLogin ? "New to WealthLens?" : "Already have an account?"}{" "}
        <Link
          className="focus-ring rounded font-semibold text-[var(--brand)] underline-offset-4 hover:underline"
          href={isLogin ? "/register" : "/login"}
        >
          {isLogin ? "Create an account" : "Sign in"}
        </Link>
      </p>
    </form>
  );
}

function Field({
  id,
  label,
  error,
  children,
}: {
  id: string;
  label: string;
  error?: string;
  children: React.ReactNode;
}) {
  return (
    <div>
      <label className="mb-2 block text-sm font-semibold" htmlFor={id}>
        {label}
      </label>
      {children}
      {error && <p className="mt-1.5 text-xs text-[var(--danger)]">{error}</p>}
    </div>
  );
}
