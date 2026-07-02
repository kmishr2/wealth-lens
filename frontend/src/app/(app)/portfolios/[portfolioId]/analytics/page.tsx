import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { HealthScoreCard } from "@/components/health-score-card";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type {
  Portfolio,
  PortfolioHealth,
  PortfolioPerformance,
  PortfolioRisk,
  PortfolioSnapshot,
} from "@/lib/types";

export const metadata: Metadata = { title: "Portfolio analytics" };

type MetricState<T> = { data: T | null; error: string };

async function loadAnalytics(
  portfolioID: string,
  requestedStart?: string,
  requestedEnd?: string,
) {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  const encodedID = encodeURIComponent(portfolioID);
  try {
    const [portfolio, snapshots] = await Promise.all([
      apiRequest<Portfolio>(`/portfolios/${encodedID}`, { accessToken }),
      apiRequest<PortfolioSnapshot[]>(`/portfolios/${encodedID}/snapshots?limit=100`, { accessToken }),
    ]);
    const daily = snapshots.filter((snapshot) => snapshot.snapshot_period === "daily");
    const availableDates = new Set(daily.map((snapshot) => snapshot.snapshot_date));
    const latest = daily[0]?.snapshot_date;
    const oldest = daily[daily.length - 1]?.snapshot_date;
    const startDate = requestedStart && availableDates.has(requestedStart) ? requestedStart : oldest;
    const endDate = requestedEnd && availableDates.has(requestedEnd) ? requestedEnd : latest;

    if (!startDate || !endDate || startDate >= endDate) {
      return { portfolio, daily, startDate, endDate, performance: emptyMetric<PortfolioPerformance>(), risk: emptyMetric<PortfolioRisk>(), health: emptyMetric<PortfolioHealth>() };
    }

    const query = `start_date=${encodeURIComponent(startDate)}&end_date=${encodeURIComponent(endDate)}`;
    const performance = await metricRequest<PortfolioPerformance>(`/portfolios/${encodedID}/performance?${query}`, accessToken);
    const risk = daily.length >= 3
      ? await metricRequest<PortfolioRisk>(`/portfolios/${encodedID}/risk?${query}&periods_per_year=252`, accessToken)
      : { data: null, error: "Risk metrics require at least three daily snapshots." };
    const health = await metricRequest<PortfolioHealth>(`/portfolios/${encodedID}/health-score`, accessToken, {
      method: "POST",
      body: JSON.stringify({ as_of_date: endDate, risk_profile: "moderate" }),
    });
    return { portfolio, daily, startDate, endDate, performance, risk, health };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) redirect(`/auth/refresh?next=${encodeURIComponent(`/portfolios/${portfolioID}/analytics`)}`);
      redirect("/login");
    }
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
}

async function metricRequest<T>(path: string, accessToken: string, options: RequestInit = {}): Promise<MetricState<T>> {
  try {
    return { data: await apiRequest<T>(path, { ...options, accessToken }), error: "" };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) throw error;
    return { data: null, error: error instanceof Error ? error.message : "Metric unavailable." };
  }
}

function emptyMetric<T>(): MetricState<T> {
  return { data: null, error: "Select two distinct daily snapshot dates." };
}

export default async function AnalyticsPage({
  params,
  searchParams,
}: {
  params: Promise<{ portfolioId: string }>;
  searchParams: Promise<{ start?: string; end?: string }>;
}) {
  const { portfolioId } = await params;
  const requested = await searchParams;
  const { portfolio, daily, startDate, endDate, performance, risk, health } = await loadAnalytics(
    portfolioId,
    requested.start,
    requested.end,
  );

  return (
    <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
      <Link className="focus-ring inline-flex rounded-lg text-sm font-semibold text-[var(--brand)] hover:underline" href={`/portfolios/${portfolio.id}`}>
        ← {portfolio.name}
      </Link>
      <div className="mt-7 flex flex-col justify-between gap-6 border-b border-[var(--line)] pb-9 lg:flex-row lg:items-end">
        <div>
          <p className="eyebrow">Snapshot-backed</p>
          <h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">Performance & risk</h1>
          <p className="mt-4 max-w-2xl leading-7 text-[var(--muted)]">
            Every value below comes from immutable snapshots and disclosed backend formulas. Missing history is shown rather than estimated.
          </p>
        </div>
        <form className="grid gap-3 rounded-2xl border border-[var(--line)] bg-[var(--surface)] p-4 sm:grid-cols-[1fr_1fr_auto]" method="GET">
          <DateField label="Start" name="start" value={startDate ?? ""} />
          <DateField label="End" name="end" value={endDate ?? ""} />
          <button className="focus-ring self-end rounded-xl bg-[var(--brand)] px-4 py-2.5 text-sm font-semibold text-white" type="submit">Apply</button>
        </form>
      </div>

      {daily.length < 2 ? (
        <Unavailable message="Analytics require at least two daily snapshots. Run the daily snapshot job after recording transactions and prices." />
      ) : (
        <div className="mt-9 space-y-9">
          <section aria-labelledby="performance-title">
            <SectionHeading eyebrow="Returns" title="Performance" id="performance-title" />
            {performance.data ? (
              <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {performance.data.currency_returns.map((item) => (
                  <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6" key={item.currency}>
                    <div className="flex items-center justify-between"><p className="eyebrow">{item.currency}</p><span className="text-xs text-[var(--muted)]">{item.cash_flow_count} cash flows</span></div>
                    <Metric label="Profit / loss" value={formatMoney(item.profit_loss, item.currency)} />
                    <div className="mt-5 grid grid-cols-2 gap-4 border-t border-[var(--line)] pt-4">
                      <SmallMetric label="CAGR" value={formatPercent(item.cagr)} />
                      <SmallMetric label="XIRR" value={formatPercent(item.xirr)} />
                    </div>
                    <p className="mt-4 text-xs text-[var(--muted)]">Net external flow: {formatMoney(item.net_external_cash_flow, item.currency)}</p>
                  </article>
                ))}
              </div>
            ) : <Unavailable message={performance.error} />}
          </section>

          <section aria-labelledby="risk-title">
            <SectionHeading eyebrow="Historical observations" title="Risk" id="risk-title" />
            {risk.data ? (
              <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {risk.data.currency_risk.map((item) => (
                  <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6" key={item.currency}>
                    <p className="eyebrow">{item.currency}</p>
                    <div className="mt-5 grid grid-cols-2 gap-5">
                      <SmallMetric label="Volatility" value={formatPercent(item.annualized_volatility)} />
                      <SmallMetric label="Max drawdown" value={formatPercent(item.maximum_drawdown)} />
                    </div>
                    <p className="mt-5 border-t border-[var(--line)] pt-4 text-xs leading-5 text-[var(--muted)]">
                      {item.observation_count} observations · peak {item.peak_date} · trough {item.trough_date}
                    </p>
                  </article>
                ))}
              </div>
            ) : <Unavailable message={risk.error} />}
          </section>

          <section aria-labelledby="health-title">
            <SectionHeading eyebrow="Rule-based score" title="Portfolio health" id="health-title" />
            {health.data ? (
              <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {health.data.scores.map((score) => <HealthScoreCard key={score.currency} score={score} />)}
              </div>
            ) : <Unavailable message={health.error} />}
          </section>
        </div>
      )}
    </main>
  );
}

function DateField({ label, name, value }: { label: string; name: string; value: string }) {
  return <label className="text-xs font-semibold text-[var(--muted)]">{label}<input className="focus-ring mt-1 block w-full rounded-lg border border-[var(--line)] bg-white px-3 py-2 text-sm text-[var(--ink)]" defaultValue={value} name={name} type="date" /></label>;
}

function SectionHeading({ eyebrow, title, id }: { eyebrow: string; title: string; id: string }) {
  return <div><p className="eyebrow">{eyebrow}</p><h2 className="mt-2 text-2xl font-semibold tracking-[-0.03em]" id={id}>{title}</h2></div>;
}

function Metric({ label, value }: { label: string; value: string }) {
  return <div className="mt-5"><p className="text-xs font-semibold uppercase tracking-[0.1em] text-[var(--muted)]">{label}</p><p className="mt-1 text-3xl font-semibold tracking-[-0.04em]">{value}</p></div>;
}

function SmallMetric({ label, value }: { label: string; value: string }) {
  return <div><p className="text-xs text-[var(--muted)]">{label}</p><p className="mt-1 font-mono text-lg font-semibold">{value}</p></div>;
}

function Unavailable({ message }: { message: string }) {
  return <div className="mt-5 rounded-2xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-6 text-sm leading-6 text-[var(--muted)]">{message}</div>;
}

function formatMoney(value: string, currency: string) {
  return new Intl.NumberFormat("en-IN", { style: "currency", currency, maximumFractionDigits: 2 }).format(Number(value));
}

function formatPercent(value: string) {
  return `${new Intl.NumberFormat("en-IN", { maximumFractionDigits: 2 }).format(Number(value))}%`;
}
