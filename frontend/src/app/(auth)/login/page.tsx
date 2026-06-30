import type { Metadata } from "next";
import { AuthForm } from "@/components/auth-form";
import { AuthShell } from "@/components/auth-shell";

export const metadata: Metadata = { title: "Sign in" };

export default async function LoginPage({
  searchParams,
}: {
  searchParams: Promise<{ reason?: string }>;
}) {
  const reason = (await searchParams).reason;
  const notice =
    reason === "session-expired"
      ? "Your session expired. Sign in again to continue."
      : reason === "backend-unavailable"
        ? "The backend is temporarily unavailable. Your session was preserved."
        : undefined;

  return (
    <AuthShell
      eyebrow="Welcome back"
      title="See your portfolio clearly."
      description="Sign in to continue to your transaction-ledger portfolio workspace."
    >
      <AuthForm mode="login" notice={notice} />
    </AuthShell>
  );
}
