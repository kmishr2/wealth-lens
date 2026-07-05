import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { HealthScoreCard } from "@/components/health-score-card";
import { CreateSnapshotForm } from "@/components/create-snapshot-form";
import { RebalancingForm } from "@/components/rebalancing-form";
import { SnapshotValueChart } from "@/components/snapshot-value-chart";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type {
  Benchmark,
  BenchmarkBeta,
  BenchmarkComparison,
  ContributionAnalysis,
  Portfolio,
  PortfolioAllocation,
  PortfolioConcentration,
  PortfolioDiversificationAlerts,
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
  requestedBenchmark?: string,
) {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  const encodedID = encodeURIComponent(portfolioID);
  try {
    const [portfolio, snapshots, benchmarks] = await Promise.all([
      apiRequest<Portfolio>(`/portfolios/${encodedID}`, { accessToken }),
      apiRequest<PortfolioSnapshot[]>(`/portfolios/${encodedID}/snapshots?limit=100`, { accessToken }),
      apiRequest<Benchmark[]>("/benchmarks?limit=100", { accessToken }),
    ]);
    const daily = snapshots.filter((snapshot) => snapshot.snapshot_period === "daily");
    const availableDates = new Set(daily.map((snapshot) => snapshot.snapshot_date));
    const latest = daily[0]?.snapshot_date;
    const oldest = daily[daily.length - 1]?.snapshot_date;
    const startDate = requestedStart && availableDates.has(requestedStart) ? requestedStart : oldest;
    const endDate = requestedEnd && availableDates.has(requestedEnd) ? requestedEnd : latest;
    const selectedBenchmark = benchmarks.find((benchmark) => benchmark.id === requestedBenchmark) ?? benchmarks[0] ?? null;
    const [allocation, concentration, alerts] = await Promise.all([
      metricRequest<PortfolioAllocation>(`/portfolios/${encodedID}/allocation`, accessToken),
      metricRequest<PortfolioConcentration>(`/portfolios/${encodedID}/concentration`, accessToken),
      metricRequest<PortfolioDiversificationAlerts>(`/portfolios/${encodedID}/diversification-alerts`, accessToken),
    ]);

    if (!startDate || !endDate || startDate >= endDate) {
      return { portfolio, daily, benchmarks, selectedBenchmark, startDate, endDate, allocation, concentration, alerts, performance: emptyMetric<PortfolioPerformance>(), contribution: emptyMetric<ContributionAnalysis>(), risk: emptyMetric<PortfolioRisk>(), health: emptyMetric<PortfolioHealth>(), comparison: emptyMetric<BenchmarkComparison>(), beta: emptyMetric<BenchmarkBeta>() };
    }

    const query = `start_date=${encodeURIComponent(startDate)}&end_date=${encodeURIComponent(endDate)}`;
    const performance = await metricRequest<PortfolioPerformance>(`/portfolios/${encodedID}/performance?${query}`, accessToken);
    const contribution = await metricRequest<ContributionAnalysis>(`/portfolios/${encodedID}/contributions?${query}&currency=${encodeURIComponent(portfolio.base_currency)}`, accessToken);
    const risk = daily.length >= 3
      ? await metricRequest<PortfolioRisk>(`/portfolios/${encodedID}/risk?${query}&periods_per_year=252`, accessToken)
      : { data: null, error: "Risk metrics require at least three daily snapshots." };
    const health = await metricRequest<PortfolioHealth>(`/portfolios/${encodedID}/health-score`, accessToken, {
      method: "POST",
      body: JSON.stringify({ as_of_date: endDate, risk_profile: "moderate" }),
    });
    let comparison = { data: null, error: "Choose a benchmark." } as MetricState<BenchmarkComparison>;
    let beta = { data: null, error: "Choose a benchmark." } as MetricState<BenchmarkBeta>;
    if (selectedBenchmark) {
      const benchmarkPath = `/portfolios/${encodedID}/benchmarks/${encodeURIComponent(selectedBenchmark.id)}`;
      const benchmarkQuery = `${query}&currency=${encodeURIComponent(selectedBenchmark.currency)}`;
      [comparison, beta] = await Promise.all([
        metricRequest<BenchmarkComparison>(`${benchmarkPath}/comparison?${benchmarkQuery}`, accessToken),
        metricRequest<BenchmarkBeta>(`${benchmarkPath}/beta?${benchmarkQuery}`, accessToken),
      ]);
    }
    return { portfolio, daily, benchmarks, selectedBenchmark, startDate, endDate, allocation, concentration, alerts, performance, contribution, risk, health, comparison, beta };
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
  searchParams: Promise<{ start?: string; end?: string; benchmark?: string }>;
}) {
  const { portfolioId } = await params;
  const requested = await searchParams;
  const { portfolio, daily, benchmarks, selectedBenchmark, startDate, endDate, allocation, concentration, alerts, performance, contribution, risk, health, comparison, beta } = await loadAnalytics(
    portfolioId,
    requested.start,
    requested.end,
    requested.benchmark,
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
        <form className="grid gap-3 rounded-2xl border border-[var(--line)] bg-[var(--surface)] p-4 sm:grid-cols-[1fr_1fr_1.2fr_auto]" method="GET">
          <DateField label="Start" name="start" value={startDate ?? ""} />
          <DateField label="End" name="end" value={endDate ?? ""} />
          <label className="text-xs font-semibold text-[var(--muted)]">
            Benchmark
            <select className="focus-ring mt-1 block w-full rounded-lg border border-[var(--line)] bg-white px-3 py-2 text-sm text-[var(--ink)]" defaultValue={selectedBenchmark?.id ?? ""} name="benchmark">
              <option value="">None available</option>
              {benchmarks.map((benchmark) => <option key={benchmark.id} value={benchmark.id}>{benchmark.code} · {benchmark.currency}</option>)}
            </select>
          </label>
          <button className="focus-ring self-end rounded-xl bg-[var(--brand)] px-4 py-2.5 text-sm font-semibold text-white" type="submit">Apply</button>
        </form>
      </div>

      <section className="mt-7 rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6" aria-labelledby="snapshot-create-title">
        <p className="eyebrow">Manual history capture</p>
        <h2 className="mt-2 text-xl font-semibold" id="snapshot-create-title">Create a daily snapshot</h2>
        <p className="mt-2 max-w-3xl text-sm leading-6 text-[var(--muted)]">
          This derives an immutable as-of record from ledger events and prices available on the selected date. Repeating an existing date returns the original snapshot without changing it.
        </p>
        <CreateSnapshotForm portfolioID={portfolio.id} />
      </section>

      {daily.length < 2 ? (
        <Unavailable message="Analytics require at least two daily snapshots. Run the daily snapshot job after recording transactions and prices." />
      ) : (
        <div className="mt-9 space-y-9">
          <section aria-labelledby="value-history-title">
            <SectionHeading eyebrow="Daily immutable snapshots" title="Value history" id="value-history-title" />
            <SnapshotValueChart snapshots={daily} />
          </section>

          <section aria-labelledby="diversification-title">
            <SectionHeading eyebrow="Current valued assets" title="Concentration & diversification" id="diversification-title" />
            {concentration.data ? (
              <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {concentration.data.currencies.map((item) => {
                  const alert = alerts.data?.alerts.find((candidate) => candidate.currency === item.currency);
                  return (
                    <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6" key={item.currency}>
                      <div className="flex items-center justify-between gap-4">
                        <p className="eyebrow">{item.currency}</p>
                        {alert && <SeverityBadge severity={alert.severity} />}
                      </div>
                      <div className="mt-5 grid grid-cols-2 gap-5">
                        <SmallMetric label="Largest asset" value={formatPercent(item.largest_asset_percentage)} />
                        <SmallMetric label="Effective assets" value={formatNumber(item.effective_asset_count)} />
                        <SmallMetric label="Holdings" value={String(item.asset_count)} />
                        <SmallMetric label="HHI" value={formatNumber(item.herfindahl_hirschman_index)} />
                      </div>
                      {alert && alert.conditions.length > 0 && (
                        <ul className="mt-5 space-y-1 border-t border-[var(--line)] pt-4 text-xs leading-5 text-[var(--muted)]">
                          {alert.conditions.map((condition) => <li key={condition}>• {condition}</li>)}
                        </ul>
                      )}
                    </article>
                  );
                })}
              </div>
            ) : <Unavailable message={concentration.error} />}
          </section>

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

          <section aria-labelledby="contribution-title">
            <SectionHeading eyebrow="Value attribution" title="Contributions & growth" id="contribution-title" />
            {contribution.data ? (
              <div className="mt-5 rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6">
                <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-4">
                  <SmallMetric label="Deposits" value={formatMoney(contribution.data.contributions, contribution.data.currency)} />
                  <SmallMetric label="Withdrawals" value={formatMoney(contribution.data.withdrawals, contribution.data.currency)} />
                  <SmallMetric label="Net contributions" value={formatMoney(contribution.data.net_contributions, contribution.data.currency)} />
                  <SmallMetric label="Investment growth" value={formatMoney(contribution.data.investment_growth, contribution.data.currency)} />
                </div>
                {contribution.data.monthly_buckets.length > 0 && (
                  <div className="mt-6 overflow-x-auto border-t border-[var(--line)] pt-5">
                    <table className="w-full min-w-150 text-left text-sm"><thead className="text-xs uppercase tracking-wide text-[var(--muted)]"><tr><th className="pb-3">Month</th><th className="pb-3">Deposits</th><th className="pb-3">Withdrawals</th><th className="pb-3">Net</th><th className="pb-3">Events</th></tr></thead>
                      <tbody className="divide-y divide-[var(--line)]">{contribution.data.monthly_buckets.map((bucket) => <tr key={bucket.month}><td className="py-3 font-semibold">{bucket.month}</td><td>{formatMoney(bucket.contributions, contribution.data!.currency)}</td><td>{formatMoney(bucket.withdrawals, contribution.data!.currency)}</td><td>{formatMoney(bucket.net_contributions, contribution.data!.currency)}</td><td>{bucket.event_count}</td></tr>)}</tbody>
                    </table>
                  </div>
                )}
              </div>
            ) : <Unavailable message={contribution.error} />}
          </section>

          <section aria-labelledby="health-title">
            <SectionHeading eyebrow="Rule-based score" title="Portfolio health" id="health-title" />
            {health.data ? (
              <div className="mt-5 grid gap-4 md:grid-cols-2 xl:grid-cols-3">
                {health.data.scores.map((score) => <HealthScoreCard key={score.currency} score={score} />)}
              </div>
            ) : <Unavailable message={health.error} />}
          </section>

          <section aria-labelledby="benchmark-title">
            <div className="flex flex-wrap items-end justify-between gap-4">
              <SectionHeading eyebrow="Explicit reference series" title="Benchmark comparison" id="benchmark-title" />
              <Link className="focus-ring rounded-lg text-sm font-semibold text-[var(--brand)] hover:underline" href="/benchmarks">Manage benchmark data →</Link>
            </div>
            {comparison.data ? (
              <div className="mt-5 grid gap-4 lg:grid-cols-[1.2fr_0.8fr]">
                <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6">
                  <div className="flex items-center justify-between gap-4"><div><p className="eyebrow">{comparison.data.benchmark_code}</p><h3 className="mt-2 text-xl font-semibold">{comparison.data.benchmark_name}</h3></div><span className="rounded-full border border-[var(--line)] px-3 py-1 text-xs font-bold text-[var(--muted)]">{comparison.data.currency}</span></div>
                  <div className="mt-6 grid gap-5 sm:grid-cols-2">
                    <ComparisonColumn title="Portfolio" total={comparison.data.portfolio_total_return} cagr={comparison.data.portfolio_cagr} />
                    <ComparisonColumn title="Benchmark" total={comparison.data.benchmark_total_return} cagr={comparison.data.benchmark_cagr} />
                  </div>
                  <div className="mt-6 grid grid-cols-2 gap-4 border-t border-[var(--line)] pt-5">
                    <SmallMetric label="Excess return" value={formatPercent(comparison.data.excess_total_return)} />
                    <SmallMetric label="Excess CAGR" value={formatPercent(comparison.data.excess_cagr)} />
                  </div>
                </article>
                {beta.data ? (
                  <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6">
                    <p className="eyebrow">Historical co-movement</p>
                    <Metric label="Beta" value={formatNumber(beta.data.beta)} />
                    <p className="mt-5 border-t border-[var(--line)] pt-4 text-xs leading-5 text-[var(--muted)]">{beta.data.aligned_observation_count} aligned observations · {beta.data.paired_return_count} return pairs</p>
                  </article>
                ) : <Unavailable message={beta.error} />}
              </div>
            ) : <Unavailable message={comparison.error || "Choose a benchmark with exact observations on both selected dates."} />}
          </section>

          <section aria-labelledby="rebalancing-title">
            <div className="sr-only" id="rebalancing-title">Rebalancing</div>
            {allocation.data ? <RebalancingForm portfolioID={portfolio.id} allocations={allocation.data.asset_class_allocations} /> : <Unavailable message={allocation.error} />}
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

function ComparisonColumn({ title, total, cagr }: { title: string; total: string; cagr: string }) {
  return <div className="rounded-2xl bg-[#f1f2ed] p-5"><p className="text-sm font-semibold">{title}</p><div className="mt-4 grid grid-cols-2 gap-3"><SmallMetric label="Total return" value={formatPercent(total)} /><SmallMetric label="CAGR" value={formatPercent(cagr)} /></div></div>;
}

function SeverityBadge({ severity }: { severity: "none" | "notice" | "warning" | "critical" }) {
  const styles = { none: "bg-[#edf7f1] text-[var(--brand)]", notice: "bg-[#fff8e8] text-[#76551f]", warning: "bg-[#fff0d8] text-[#8a4b16]", critical: "bg-[#fff4f2] text-[var(--danger)]" };
  return <span className={`rounded-full px-3 py-1 text-xs font-bold capitalize ${styles[severity]}`}>{severity}</span>;
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

function formatNumber(value: string) {
  return new Intl.NumberFormat("en-IN", { maximumFractionDigits: 2 }).format(Number(value));
}
