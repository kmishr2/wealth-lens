import type { Metadata } from "next";
import { ProfileForm } from "@/components/profile-form";
import { requireUser } from "@/lib/auth";

export const metadata: Metadata = { title: "Profile settings" };

export default async function ProfilePage() {
  const user = await requireUser("/profile");

  return (
    <main className="mx-auto max-w-4xl px-6 py-10 lg:px-10 lg:py-14">
      <p className="eyebrow">Account</p>
      <h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">
        Profile settings
      </h1>
      <p className="mt-4 max-w-2xl leading-7 text-[var(--muted)]">
        These defaults identify you and prefill new financial records. Existing
        portfolios and accounts keep their own currencies.
      </p>

      <div className="mt-9 grid gap-6 lg:grid-cols-[minmax(0,1fr)_280px]">
        <section className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6 shadow-[var(--shadow)] sm:p-8">
          <h2 className="text-2xl font-semibold tracking-[-0.03em]">
            Personal defaults
          </h2>
          <ProfileForm user={user} />
        </section>

        <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6">
          <p className="text-xs font-bold uppercase tracking-[0.12em] text-[var(--muted)]">
            Sign-in email
          </p>
          <p className="mt-2 break-words font-semibold">{user.email}</p>
          <p className="mt-4 border-t border-[var(--line)] pt-4 text-sm leading-6 text-[var(--muted)]">
            Email changes are not supported. Your timezone controls how dated
            records and scheduled jobs are interpreted.
          </p>
        </aside>
      </div>
    </main>
  );
}
