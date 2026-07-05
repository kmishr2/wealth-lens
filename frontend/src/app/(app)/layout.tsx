import Link from "next/link";
import { logoutAction } from "@/app/actions/auth";
import { BrandMark } from "@/components/auth-shell";
import { requireUser } from "@/lib/auth";

export default async function AppLayout({
  children,
}: Readonly<{ children: React.ReactNode }>) {
  const user = await requireUser();

  return (
    <div className="min-h-screen">
      <header className="border-b border-[var(--line)] bg-[var(--surface)]/90 backdrop-blur">
        <div className="mx-auto flex max-w-7xl items-center justify-between px-6 py-4 lg:px-10">
          <Link className="focus-ring rounded-xl" href="/dashboard">
            <BrandMark />
          </Link>
          <div className="flex items-center gap-4">
            <nav className="flex items-center gap-1" aria-label="Primary navigation">
              <Link className="focus-ring hidden rounded-lg px-3 py-2 text-sm font-semibold text-[var(--muted)] hover:bg-[var(--brand-soft)] hover:text-[var(--brand)] sm:block" href="/dashboard">Portfolios</Link>
              <Link className="focus-ring rounded-lg px-3 py-2 text-sm font-semibold text-[var(--muted)] hover:bg-[var(--brand-soft)] hover:text-[var(--brand)]" href="/assets">Assets</Link>
              <Link className="focus-ring hidden rounded-lg px-3 py-2 text-sm font-semibold text-[var(--muted)] hover:bg-[var(--brand-soft)] hover:text-[var(--brand)] lg:block" href="/benchmarks">Benchmarks</Link>
            </nav>
            <div className="hidden text-right sm:block">
              <p className="text-sm font-semibold">{user.display_name}</p>
              <p className="text-xs text-[var(--muted)]">{user.email}</p>
            </div>
            <form action={logoutAction}>
              <button
                className="focus-ring rounded-xl border border-[var(--line)] bg-white px-4 py-2 text-sm font-semibold transition hover:border-[#b9c3bd] hover:bg-[#f8f8f4]"
                type="submit"
              >
                Sign out
              </button>
            </form>
          </div>
        </div>
      </header>
      {children}
    </div>
  );
}
