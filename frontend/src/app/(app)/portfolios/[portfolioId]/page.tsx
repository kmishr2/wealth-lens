import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { CreateAccountForm } from "@/components/create-account-form";
import { AllocationBars } from "@/components/allocation-bars";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type {
  Account,
  HoldingsResponse,
  Portfolio,
  PortfolioAllocation,
  PortfolioValuation,
} from "@/lib/types";

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
    const [portfolio, accounts, holdings, valuation, allocation] = await Promise.all([
      apiRequest<Portfolio>(`/portfolios/${encodedID}`, { accessToken }),
      apiRequest<Account[]>(`/portfolios/${encodedID}/accounts?limit=100`, {
        accessToken,
      }),
      apiRequest<HoldingsResponse>(`/portfolios/${encodedID}/holdings`, { accessToken }),
      apiRequest<PortfolioValuation>(`/portfolios/${encodedID}/valuation`, { accessToken }),
      apiRequest<PortfolioAllocation>(`/portfolios/${encodedID}/allocation`, { accessToken }),
    ]);
    return { portfolio, accounts, holdings, valuation, allocation };
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
  const { portfolio, accounts, holdings, valuation, allocation } =
    await loadPortfolioWorkspace(portfolioId);

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
        <div className="flex flex-wrap items-center gap-3">
          <Link className="focus-ring rounded-xl bg-[var(--brand)] px-4 py-3 text-sm font-semibold text-white" href={`/portfolios/${portfolio.id}/analytics`}>
            View analytics
          </Link>
          <Link className="focus-ring rounded-xl border border-[var(--line)] bg-white px-4 py-3 text-sm font-semibold text-[var(--brand)]" href={`/portfolios/${portfolio.id}/planning`}>
            Plan goals
          </Link>
          <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-5 py-4 sm:min-w-36">
          <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]">
            Accounts
          </p>
          <p className="mt-1 text-3xl font-semibold tracking-[-0.04em]">
            {accounts.length}
          </p>
          </div>
        </div>
      </div>

      <section className="mt-9" aria-labelledby="overview-title">
        <div className="flex flex-col justify-between gap-3 sm:flex-row sm:items-end">
          <div>
            <p className="eyebrow">Current position</p>
            <h2 className="mt-2 text-2xl font-semibold tracking-[-0.03em]" id="overview-title">
              Portfolio overview
            </h2>
          </div>
          <p className={`text-sm font-semibold ${valuation.is_fully_valued ? "text-[var(--brand)]" : "text-[var(--danger)]"}`}>
            {valuation.is_fully_valued ? "All holdings valued" : `${valuation.missing_prices.length} missing price${valuation.missing_prices.length === 1 ? "" : "s"}`}
          </p>
        </div>

        {valuation.total_values.length === 0 ? (
          <div className="mt-5 rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8">
            <h3 className="text-xl font-semibold">No portfolio value yet</h3>
            <p className="mt-2 max-w-2xl leading-7 text-[var(--muted)]">
              Record a deposit or asset transaction in an account. Holdings, valuation, and allocation are derived from those ledger events.
            </p>
          </div>
        ) : (
          <div className="mt-5 grid gap-5 lg:grid-cols-[0.9fr_1.1fr]">
            <div className="space-y-5">
              <div className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6 shadow-[0_6px_25px_rgba(23,32,28,0.04)]">
                <p className="text-xs font-bold uppercase tracking-[0.12em] text-[var(--muted)]">Total value</p>
                <div className="mt-4 space-y-3">
                  {valuation.total_values.map((total) => (
                    <p className="text-3xl font-semibold tracking-[-0.04em]" key={total.currency}>
                      {formatMoney(total.amount, total.currency)}
                    </p>
                  ))}
                </div>
                <p className="mt-5 border-t border-[var(--line)] pt-4 text-xs leading-5 text-[var(--muted)]">
                  {valuation.valuation_scope}
                </p>
              </div>

              <div className="rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6">
                <p className="eyebrow">Asset classes</p>
                <h3 className="mt-2 text-xl font-semibold">Allocation</h3>
                <div className="mt-5">
                  <AllocationBars items={allocation.asset_class_allocations} />
                </div>
              </div>
            </div>

            <div className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6">
              <div className="flex items-center justify-between gap-4">
                <div>
                  <p className="eyebrow">Ledger-derived</p>
                  <h3 className="mt-2 text-xl font-semibold">Holdings</h3>
                </div>
                <span className="rounded-full bg-[var(--brand-soft)] px-3 py-1 text-xs font-bold text-[var(--brand)]">
                  {holdings.asset_holdings.length} assets
                </span>
              </div>
              <div className="mt-5 divide-y divide-[var(--line)]">
                {holdings.asset_holdings.map((holding) => (
                  <div className="flex items-center justify-between gap-5 py-4 first:pt-0" key={holding.asset_id}>
                    <div className="min-w-0">
                      <p className="font-semibold">{holding.asset_symbol}</p>
                      <p className="truncate text-sm text-[var(--muted)]">{holding.asset_name}</p>
                    </div>
                    <div className="text-right">
                      <p className="font-mono text-sm font-semibold">{formatQuantity(holding.quantity)}</p>
                      <p className="mt-1 text-xs uppercase text-[var(--muted)]">{holding.asset_class}</p>
                    </div>
                  </div>
                ))}
                {holdings.cash_balances.map((cash) => (
                  <div className="flex items-center justify-between gap-5 py-4" key={`cash-${cash.currency}`}>
                    <div><p className="font-semibold">Cash</p><p className="text-sm text-[var(--muted)]">{cash.currency} balance</p></div>
                    <p className="font-mono text-sm font-semibold">{formatMoney(cash.amount, cash.currency)}</p>
                  </div>
                ))}
              </div>
              {valuation.missing_prices.length > 0 && (
                <div className="mt-5 rounded-2xl border border-[#e8c9c4] bg-[#fff4f2] p-4">
                  <p className="text-sm font-semibold text-[var(--danger)]">Missing explicit prices</p>
                  <p className="mt-1 text-sm leading-6 text-[var(--muted)]">
                    {valuation.missing_prices.map((item) => item.asset_symbol).join(", ")} are excluded from current value and allocation.
                  </p>
                </div>
              )}
            </div>
          </div>
        )}
      </section>

      <div className="mt-12 grid gap-8 border-t border-[var(--line)] pt-10 lg:grid-cols-[minmax(0,1fr)_360px]">
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

function formatMoney(value: string, currency: string) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(Number(value));
}

function formatQuantity(value: string) {
  return new Intl.NumberFormat("en-IN", { maximumFractionDigits: 6 }).format(
    Number(value),
  );
}
