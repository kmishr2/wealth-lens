import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { CreateTransactionForm } from "@/components/create-transaction-form";
import { AccountSettings } from "@/components/account-settings";
import { TransactionAuditActions } from "@/components/transaction-audit-actions";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Account, Asset, Portfolio, Transaction } from "@/lib/types";

export const metadata: Metadata = { title: "Account ledger" };

async function loadAccountLedger(portfolioID: string, accountID: string) {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  const portfolioPath = encodeURIComponent(portfolioID);
  const accountPath = encodeURIComponent(accountID);
  try {
    const [portfolio, account, allTransactions, assets] = await Promise.all([
      apiRequest<Portfolio>(`/portfolios/${portfolioPath}`, { accessToken }),
      apiRequest<Account>(`/portfolios/${portfolioPath}/accounts/${accountPath}`, { accessToken }),
      apiRequest<Transaction[]>(`/portfolios/${portfolioPath}/transactions?limit=100`, { accessToken }),
      apiRequest<Asset[]>("/assets?limit=100", { accessToken }),
    ]);
    return {
      portfolio,
      account,
      transactions: allTransactions.filter((transaction) => transaction.account_id === accountID),
      assets: assets.filter((asset) => asset.is_active && asset.currency === account.currency),
    };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) {
        redirect(`/auth/refresh?next=${encodeURIComponent(`/portfolios/${portfolioID}/accounts/${accountID}`)}`);
      }
      redirect("/login");
    }
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
}

export default async function AccountLedgerPage({
  params,
}: {
  params: Promise<{ portfolioId: string; accountId: string }>;
}) {
  const { portfolioId, accountId } = await params;
  const { portfolio, account, transactions, assets } = await loadAccountLedger(portfolioId, accountId);
  const supersededTransactionIDs = new Set(
    transactions.flatMap((transaction) =>
      [transaction.reverses_transaction_id, transaction.corrects_transaction_id].filter(
        (id): id is string => Boolean(id),
      ),
    ),
  );

  return (
    <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
      <Link
        className="focus-ring inline-flex rounded-lg text-sm font-semibold text-[var(--brand)] hover:underline"
        href={`/portfolios/${portfolio.id}`}
      >
        ← {portfolio.name}
      </Link>

      <div className="mt-7 flex flex-col justify-between gap-6 border-b border-[var(--line)] pb-9 sm:flex-row sm:items-end">
        <div>
          <div className="flex items-center gap-3">
            <p className="eyebrow">Transaction ledger</p>
            <span className="rounded-full border border-[var(--line)] bg-white px-3 py-1 text-xs font-bold text-[var(--muted)]">
              {account.currency}
            </span>
          </div>
          <h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">{account.name}</h1>
          <p className="mt-4 max-w-2xl leading-7 text-[var(--muted)]">
            {account.institution_name || "Institution not specified"}. Every correction is recorded as a new audit event; historical entries are never edited.
          </p>
        </div>
        <div className="rounded-2xl border border-[var(--line)] bg-[var(--surface)] px-5 py-4 sm:min-w-44">
          <p className="text-xs font-semibold uppercase tracking-[0.12em] text-[var(--muted)]">Events</p>
          <p className="mt-1 text-3xl font-semibold tracking-[-0.04em]">{transactions.length}</p>
        </div>
      </div>

      <div className="mt-9 grid gap-8 lg:grid-cols-[minmax(0,1fr)_380px]">
        <section aria-labelledby="ledger-title">
          <div className="mb-5 flex items-end justify-between gap-4">
            <div>
              <p className="eyebrow">History</p>
              <h2 className="mt-2 text-2xl font-semibold tracking-[-0.03em]" id="ledger-title">Recorded events</h2>
            </div>
            <p className="text-sm text-[var(--muted)]">Newest first</p>
          </div>
          {transactions.length === 0 ? (
            <div className="grid min-h-72 place-items-center rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-center">
              <div className="max-w-md">
                <h3 className="text-2xl font-semibold tracking-[-0.03em]">No ledger events yet</h3>
                <p className="mt-3 leading-7 text-[var(--muted)]">Record the opening deposit before purchases so cash and holdings remain explainable.</p>
              </div>
            </div>
          ) : (
            <div className="space-y-3">
              {transactions.map((transaction) => (
                <article className="rounded-2xl border border-[var(--line)] bg-[var(--surface-strong)] p-5" key={transaction.id}>
                  <div className="flex flex-wrap items-start justify-between gap-4">
                    <div>
                      <p className="text-sm font-bold uppercase tracking-[0.08em] text-[var(--brand)]">{transaction.transaction_type}</p>
                      <p className="mt-1 font-semibold">{transaction.description || "No description"}</p>
                    </div>
                    <time className="text-sm text-[var(--muted)]" dateTime={transaction.occurred_at}>
                      {new Intl.DateTimeFormat("en-IN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(transaction.occurred_at))}
                    </time>
                  </div>
                  <div className="mt-4 grid gap-2 border-t border-[var(--line)] pt-4 sm:grid-cols-2">
                    {transaction.entries.map((entry) => (
                      <div className="flex items-center justify-between gap-3 text-sm" key={entry.id}>
                        <span className="capitalize text-[var(--muted)]">{entry.entry_kind}</span>
                        <span className="font-mono font-semibold">
                          {entry.quantity ? `${entry.quantity} units` : `${entry.amount} ${entry.currency}`}
                        </span>
                      </div>
                    ))}
                  </div>
                  <TransactionAuditActions
                    accountID={account.id}
                    assets={assets}
                    currency={account.currency}
                    portfolioID={portfolio.id}
                    superseded={supersededTransactionIDs.has(transaction.id)}
                    transaction={transaction}
                  />
                </article>
              ))}
            </div>
          )}
        </section>

        <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6 shadow-[var(--shadow)] sm:p-7">
          <p className="eyebrow">New ledger event</p>
          <h2 className="mt-2 text-2xl font-semibold tracking-[-0.03em]">Record a transaction</h2>
          <p className="mt-3 text-sm leading-6 text-[var(--muted)]">
            Enter positive amounts. WealthLens applies ledger signs based on the event type.
          </p>
          <CreateTransactionForm portfolioID={portfolio.id} accountID={account.id} currency={account.currency} assets={assets} />
        </aside>
      </div>

      <section className="mt-12 border-t border-[var(--line)] pt-10" aria-labelledby="account-settings-title">
        <p className="eyebrow">Administration</p>
        <h2 className="mt-2 text-2xl font-semibold tracking-[-0.03em]" id="account-settings-title">
          Account settings
        </h2>
        <AccountSettings account={account} />
      </section>
    </main>
  );
}
