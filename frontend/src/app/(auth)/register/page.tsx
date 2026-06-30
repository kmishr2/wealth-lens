import type { Metadata } from "next";
import { AuthForm } from "@/components/auth-form";
import { AuthShell } from "@/components/auth-shell";

export const metadata: Metadata = { title: "Create account" };

export default function RegisterPage() {
  return (
    <AuthShell
      eyebrow="Start with the ledger"
      title="Build a transparent record."
      description="Create a workspace for deterministic portfolio tracking. WealthLens does not provide financial advice."
    >
      <AuthForm mode="register" />
    </AuthShell>
  );
}
