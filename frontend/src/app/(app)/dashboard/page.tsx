import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";
import { CreatePortfolioForm } from "@/components/create-portfolio-form";
import { apiRequest, ApiError } from "@/lib/api";
import { requireUser } from "@/lib/auth";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Portfolio } from "@/lib/types";

export const metadata: Metadata = { title: "Portfolios" };

async function listPortfolios(): Promise<Portfolio[]> {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  try {
    return await apiRequest<Portfolio[]>("/portfolios?limit=100", {
      accessToken,
    });
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) {
        redirect("/auth/refresh?next=/dashboard");
      }
      redirect("/login");
    }
    throw error;
  }
}

export default async function DashboardPage() {
  const [user, portfolios] = await Promise.all([
    requireUser(),
    listPortfolios(),
  ]);

  return (
    <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
      <div className="flex flex-col justify-between gap-6 border-b border-[var(--line)] pb-9 sm:flex-row sm:items-end">
        <div>
          <p className="eyebrow">Portfolio workspace</p>
          <h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">
            Your portfolios
          </h1>
          <p className="mt-4 max-w-2xl leading-7 text-[var(--muted)]">
            Each portfolio is derived from its transaction ledger. Values and
            analytics will appear only after you record financial events.
          </p>
        </div>
        <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-5 py-4 sm:min-w-44">
          <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]">
            Portfolios
          </p>
          <p className="mt-1 text-3xl font-semibold tracking-[-0.04em]">
            {portfolios.length}
          </p>
        </div>
      </div>

      <div className="mt-9 grid gap-8 lg:grid-cols-[minmax(0,1fr)_360px]">
        <section aria-labelledby="portfolio-list-title">
          <h2 className="sr-only" id="portfolio-list-title">
            Portfolio list
          </h2>
          {portfolios.length === 0 ? (
            <div className="grid min-h-80 place-items-center rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-center">
              <div className="max-w-md">
                <span className="mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-[var(--brand-soft)] text-xl font-bold text-[var(--brand)]">
                  0
                </span>
                <h3 className="mt-5 text-2xl font-semibold tracking-[-0.03em]">
                  No portfolios yet
                </h3>
                <p className="mt-3 leading-7 text-[var(--muted)]">
                  Create your first portfolio, then add an account and record
                  its opening contribution as a ledger transaction.
                </p>
              </div>
            </div>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2">
              {portfolios.map((portfolio, index) => (
                <Link
                  className="focus-ring group rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6 shadow-[0_6px_25px_rgba(23,32,28,0.04)] transition hover:-translate-y-0.5 hover:shadow-[var(--shadow)]"
                  href={`/portfolios/${portfolio.id}`}
                  key={portfolio.id}
                >
                  <div className="flex items-start justify-between gap-4">
                    <span className="grid h-11 w-11 place-items-center rounded-xl bg-[var(--brand-soft)] font-bold text-[var(--brand)]">
                      {String(index + 1).padStart(2, "0")}
                    </span>
                    <span className="rounded-full border border-[var(--line)] px-3 py-1 text-xs font-bold tracking-wide text-[var(--muted)]">
                      {portfolio.base_currency}
                    </span>
                  </div>
                  <h3 className="mt-6 text-xl font-semibold tracking-[-0.025em]">
                    {portfolio.name}
                  </h3>
                  <p className="mt-2 min-h-12 text-sm leading-6 text-[var(--muted)]">
                    {portfolio.description || "No description provided."}
                  </p>
                  <div className="mt-6 flex items-center justify-between border-t border-[var(--line)] pt-4 text-xs text-[var(--muted)]">
                    <span>Ledger-backed</span>
                    <span className="font-semibold text-[var(--brand)]">
                      Setup next →
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </section>

        <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6 shadow-[var(--shadow)] sm:p-7">
          <p className="eyebrow">New portfolio</p>
          <h2 className="mt-2 text-2xl font-semibold tracking-[-0.03em]">
            Start a ledger
          </h2>
          <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
            The base currency labels the portfolio. WealthLens does not perform
            implicit currency conversion.
          </p>
          <CreatePortfolioForm defaultCurrency={user.base_currency} />
        </aside>
      </div>
    </main>
  );
}
