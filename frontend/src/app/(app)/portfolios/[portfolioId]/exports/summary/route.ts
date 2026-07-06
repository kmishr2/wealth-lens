import { redirect } from "next/navigation";
import { apiRequest, ApiError } from "@/lib/api";
import { safeFilename } from "@/lib/csv";
import { buildPortfolioSummaryCSV } from "@/lib/portfolio-report";
import { getAccessToken, getRefreshToken, setSession } from "@/lib/session";
import type { AuthResponse, HoldingsResponse, Portfolio, PortfolioAllocation, PortfolioValuation } from "@/lib/types";

export async function GET(_request: Request, { params }: { params: Promise<{ portfolioId: string }> }) {
  const { portfolioId } = await params;
  let accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  let refreshInFlight: Promise<string> | null = null;

  async function authenticatedRequest<T>(path: string): Promise<T> {
    try {
      return await apiRequest<T>(path, { accessToken });
    } catch (error) {
      if (!(error instanceof ApiError) || error.status !== 401) throw error;
      if (!refreshInFlight) refreshInFlight = refreshSession().finally(() => { refreshInFlight = null; });
      accessToken = await refreshInFlight;
      return apiRequest<T>(path, { accessToken });
    }
  }

  async function refreshSession() {
    const refreshToken = await getRefreshToken();
    if (!refreshToken) redirect("/login");
    try {
      const auth = await apiRequest<AuthResponse>("/auth/refresh", { method: "POST", body: JSON.stringify({ refresh_token: refreshToken }) });
      await setSession(auth);
      return auth.access_token;
    } catch (error) {
      if (error instanceof ApiError && error.status === 401) redirect("/login");
      throw error;
    }
  }

  const encodedID = encodeURIComponent(portfolioId);
  try {
    const [portfolio, holdings, valuation, allocation] = await Promise.all([
      authenticatedRequest<Portfolio>(`/portfolios/${encodedID}`),
      authenticatedRequest<HoldingsResponse>(`/portfolios/${encodedID}/holdings`),
      authenticatedRequest<PortfolioValuation>(`/portfolios/${encodedID}/valuation`),
      authenticatedRequest<PortfolioAllocation>(`/portfolios/${encodedID}/allocation`),
    ]);
    const csv = buildPortfolioSummaryCSV(portfolio, holdings, valuation, allocation);
    return new Response(`\uFEFF${csv}`, { headers: {
      "Cache-Control": "no-store",
      "Content-Disposition": `attachment; filename="${safeFilename(portfolio.name)}-summary.csv"`,
      "Content-Type": "text/csv; charset=utf-8",
      "X-Content-Type-Options": "nosniff",
    } });
  } catch (error) {
    if (error instanceof ApiError) return new Response(error.message, { status: error.status >= 400 && error.status < 600 ? error.status : 500 });
    throw error;
  }
}
