import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { CreateAccountForm } from "@/components/create-account-form";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Account, Portfolio } from "@/lib/types";

export const metadata: Metadata = { title: "Portfolio setup" };

const accountTypeLabels: Record<Account["account_type"], string> = {
  brokerage: "Brokerage",
  retirement: "Retirement",
  bank: "Bank",
  wallet: "Wallet",
  other: "Other",
};

async function loadPortfolioWorkspace(portfolioID: string) {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");

  const encodedID = encodeURIComponent(portfolioID);
  try {
    const [portfolio, accounts] = await Promise.all([
      apiRequest<Portfolio>(`/portfolios/${encodedID}`, { accessToken }),
      apiRequest<Account[]>(`/portfolios/${encodedID}/accounts?limit=100`, {
        accessToken,
      }),
    ]);
    return { portfolio, accounts };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) {
        redirect(
          `/auth/refresh?next=${encodeURIComponent(`/portfolios/${portfolioID}`)}`,
        );
      }
      redirect("/login");
    }
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
}

export default async function PortfolioPage({
  params,
}: {
  params: Promise<{ portfolioId: string }>;
}) {
  const { portfolioId } = await params;
  const { portfolio, accounts } = await loadPortfolioWorkspace(portfolioId);

  return (
    <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
      <Link
        className="focus-ring inline-flex rounded-lg text-sm font-semibold text-[var(--brand)] hover:underline"
        href="/dashboard"
      >
        ← All portfolios
      </Link>

      <div className="mt-7 flex flex-col justify-between gap-6 border-b border-[var(--line)] pb-9 sm:flex-row sm:items-end">
        <div>
          <div className="flex items-center gap-3">
            <p className="eyebrow">Portfolio setup</p>
            <span className="rounded-full border border-[var(--line)] bg-white px-3 py-1 text-xs font-bold tracking-wide text-[var(--muted)]">
              {portfolio.base_currency}
            </span>
          </div>
          <h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">
            {portfolio.name}
          </h1>
          <p className="mt-4 max-w-2xl leading-7 text-[var(--muted)]">
            {portfolio.description ||
              "Add the accounts that contain this portfolio’s financial events."}
          </p>
        </div>
        <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-5 py-4 sm:min-w-44">
          <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]">
            Accounts
          </p>
          <p className="mt-1 text-3xl font-semibold tracking-[-0.04em]">
            {accounts.length}
          </p>
        </div>
      </div>

      <div className="mt-9 grid gap-8 lg:grid-cols-[minmax(0,1fr)_360px]">
        <section aria-labelledby="account-list-title">
          <h2 className="sr-only" id="account-list-title">
            Account list
          </h2>
          {accounts.length === 0 ? (
            <div className="grid min-h-80 place-items-center rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-center">
              <div className="max-w-md">
                <span className="mx-auto grid h-14 w-14 place-items-center rounded-2xl bg-[var(--brand-soft)] text-xl font-bold text-[var(--brand)]">
                  0
                </span>
                <h2 className="mt-5 text-2xl font-semibold tracking-[-0.03em]">
                  No accounts yet
                </h2>
                <p className="mt-3 leading-7 text-[var(--muted)]">
                  Accounts group ledger events by broker, bank, retirement plan,
                  or wallet. Add one before recording transactions.
                </p>
              </div>
            </div>
          ) : (
            <div className="grid gap-4 sm:grid-cols-2">
              {accounts.map((account) => (
                <Link
                  className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6 shadow-[0_6px_25px_rgba(23,32,28,0.04)]"
                  href={`/portfolios/${portfolio.id}/accounts/${account.id}`}
                  key={account.id}
                >
                  <div className="flex items-start justify-between gap-4">
                    <span className="grid h-11 w-11 place-items-center rounded-xl bg-[var(--brand-soft)] text-sm font-black text-[var(--brand)]">
                      {account.name.slice(0, 2).toUpperCase()}
                    </span>
                    <span className="rounded-full border border-[var(--line)] px-3 py-1 text-xs font-bold tracking-wide text-[var(--muted)]">
                      {account.currency}
                    </span>
                  </div>
                  <h2 className="mt-6 text-xl font-semibold tracking-[-0.025em]">
                    {account.name}
                  </h2>
                  <p className="mt-2 text-sm text-[var(--muted)]">
                    {account.institution_name || "Institution not specified"}
                  </p>
                  <div className="mt-6 flex items-center justify-between border-t border-[var(--line)] pt-4 text-xs">
                    <span className="font-semibold text-[var(--muted)]">
                      {accountTypeLabels[account.account_type]}
                    </span>
                    <span className="font-semibold text-[var(--brand)]">
                      Transactions next →
                    </span>
                  </div>
                </Link>
              ))}
            </div>
          )}
        </section>

        <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6 shadow-[var(--shadow)] sm:p-7">
          <p className="eyebrow">New account</p>
          <h2 className="mt-2 text-2xl font-semibold tracking-[-0.03em]">
            Add an account
          </h2>
          <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
            This groups transactions only. WealthLens does not connect to the
            institution or execute trades.
          </p>
          <CreateAccountForm
            defaultCurrency={portfolio.base_currency}
            portfolioID={portfolio.id}
          />
        </aside>
      </div>
    </main>
  );
}
