import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import { csvCell, safeFilename, safeSpreadsheetText } from "@/lib/csv";
import type { Account, Asset, AuthResponse, Portfolio, Transaction } from "@/lib/types";

const pageSize = 200;

export async function GET(
  request: Request,
  { params }: { params: Promise<{ portfolioId: string }> },
) {
  const { portfolioId } = await params;
  const accountFilter = new URL(request.url).searchParams.get("account")?.trim() ?? "";
  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  let refreshInFlight: Promise<string> | null = null;

  async function authenticatedRequest<T>(path: string): Promise<T> {
    try {
      return await apiRequest<T>(path, { accessToken });
    } catch (error) {
      if (!(error instanceof ApiError) || error.status !== 401) throw error;
      if (!refreshInFlight) {
        refreshInFlight = refreshSession().finally(() => {
          refreshInFlight = null;
        });
      }
      accessToken = await refreshInFlight;
      return apiRequest<T>(path, { accessToken });
    }
  }

  async function refreshSession() {
    const refreshToken = await getRefreshToken();
    if (!refreshToken) redirect("/login");
    try {
      const auth = await apiRequest<AuthResponse>("/auth/refresh", {
        method: "POST",
        body: JSON.stringify({ refresh_token: refreshToken }),
      });
      await setSession(auth);
      return auth.access_token;
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) redirect("/login");
      throw error;
    }
  }

  const encodedPortfolioID = encodeURIComponent(portfolioId);
  try {
    const [portfolio, accounts, assets, transactions] = await Promise.all([
      authenticatedRequest<Portfolio>(`/portfolios/${encodedPortfolioID}`),
      fetchAll<Account>(
        (offset) =>
          authenticatedRequest<Account[]>(
            `/portfolios/${encodedPortfolioID}/accounts?limit=${pageSize}&offset=${offset}`,
          ),
      ),
      fetchAll<Asset>((offset) =>
        authenticatedRequest<Asset[]>(`/assets?limit=${pageSize}&offset=${offset}`),
      ),
      fetchAll<Transaction>((offset) =>
        authenticatedRequest<Transaction[]>(
          `/portfolios/${encodedPortfolioID}/transactions?limit=${pageSize}&offset=${offset}`,
        ),
      ),
    ]);

    if (accountFilter && !accounts.some((account) => account.id === accountFilter)) {
      return new Response("Account not found in this portfolio.", { status: 404 });
    }

    const accountNames = new Map(accounts.map((account) => [account.id, account.name]));
    const assetSymbols = new Map(assets.map((asset) => [asset.id, asset.symbol]));
    const selectedTransactions = accountFilter
      ? transactions.filter((transaction) => transaction.account_id === accountFilter)
      : transactions;
    const csv = buildTransactionCSV(selectedTransactions, accountNames, assetSymbols);
    const scopeName = accountFilter
      ? accountNames.get(accountFilter) ?? "account"
      : portfolio.name;

    return new Response(`\uFEFF${csv}`, {
      headers: {
        "Cache-Control": "no-store",
        "Content-Disposition": `attachment; filename="${safeFilename(scopeName)}-ledger.csv"`,
        "Content-Type": "text/csv; charset=utf-8",
        "X-Content-Type-Options": "nosniff",
      },
    });
  } catch (error) {
    if (error instanceof ApiError) {
      return new Response(error.message, {
        status: error.status >= 400 && error.status < 600 ? error.status : 500,
      });
    }
    throw error;
  }
}

async function fetchAll<T>(loadPage: (offset: number) => Promise<T[]>) {
  const records: T[] = [];
  let offset = 0;
  while (true) {
    const page = await loadPage(offset);
    records.push(...page);
    if (page.length < pageSize) return records;
    offset += pageSize;
  }
}

function buildTransactionCSV(
  transactions: Transaction[],
  accountNames: Map<string, string>,
  assetSymbols: Map<string, string>,
) {
  const headers = [
    "transaction_id",
    "occurred_at",
    "transaction_type",
    "account_id",
    "account_name",
    "description",
    "entry_id",
    "entry_kind",
    "asset_id",
    "asset_symbol",
    "quantity",
    "amount",
    "currency",
    "reverses_transaction_id",
    "corrects_transaction_id",
    "created_at",
  ];
  const rows = transactions.flatMap((transaction) =>
    transaction.entries.map((entry) => [
      transaction.id,
      transaction.occurred_at,
      transaction.transaction_type,
      transaction.account_id,
      safeSpreadsheetText(accountNames.get(transaction.account_id) ?? ""),
      safeSpreadsheetText(transaction.description),
      entry.id,
      entry.entry_kind,
      entry.asset_id ?? "",
      safeSpreadsheetText(entry.asset_id ? assetSymbols.get(entry.asset_id) ?? "" : ""),
      entry.quantity ?? "",
      entry.amount ?? "",
      entry.currency,
      transaction.reverses_transaction_id ?? "",
      transaction.corrects_transaction_id ?? "",
      transaction.created_at,
    ]),
  );
  return [headers, ...rows]
    .map((row) => row.map((value) => csvCell(value)).join(","))
    .join("\r\n");
}
