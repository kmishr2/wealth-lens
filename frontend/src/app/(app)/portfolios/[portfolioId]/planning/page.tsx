import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { CreateGoalForm } from "@/components/create-goal-form";
import { GoalActions } from "@/components/goal-actions";
import { SIPCalculator, WhatIfCalculator } from "@/components/planning-calculators";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Goal, MonthlyGoalSnapshot, Portfolio } from "@/lib/types";

export const metadata: Metadata = { title: "Goals & planning" };

async function loadPlanning(portfolioID: string) {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  const encodedID = encodeURIComponent(portfolioID);
  try {
    const [portfolio, goals] = await Promise.all([
      apiRequest<Portfolio>(`/portfolios/${encodedID}`, { accessToken }),
      apiRequest<Goal[]>(`/portfolios/${encodedID}/goals?limit=100`, { accessToken }),
    ]);
    const snapshots = await Promise.all(
      goals.map(async (goal) => {
        const records = await apiRequest<MonthlyGoalSnapshot[]>(
          `/portfolios/${encodedID}/goals/${encodeURIComponent(goal.id)}/monthly-snapshots?limit=1`,
          { accessToken },
        );
        return [goal.id, records[0] ?? null] as const;
      }),
    );
    return { portfolio, goals, latestByGoal: new Map(snapshots) };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) redirect(`/auth/refresh?next=${encodeURIComponent(`/portfolios/${portfolioID}/planning`)}`);
      redirect("/login");
    }
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
}

export default async function PlanningPage({ params }: { params: Promise<{ portfolioId: string }> }) {
  const { portfolioId } = await params;
  const { portfolio, goals, latestByGoal } = await loadPlanning(portfolioId);
  return (
    <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
      <Link className="focus-ring inline-flex rounded-lg text-sm font-semibold text-[var(--brand)] hover:underline" href={`/portfolios/${portfolio.id}`}>
        ← {portfolio.name}
      </Link>
      <div className="mt-7 border-b border-[var(--line)] pb-9">
        <p className="eyebrow">Explicit assumptions only</p>
        <h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">Goals & planning</h1>
        <p className="mt-4 max-w-3xl leading-7 text-[var(--muted)]">
          Track declared targets and run deterministic scenarios. Projections use your return and inflation inputs; they are not forecasts and are never stored as portfolio truth.
        </p>
      </div>

      <section className="mt-9" aria-labelledby="goals-title">
        <div><p className="eyebrow">Targets</p><h2 className="mt-2 text-2xl font-semibold" id="goals-title">Financial goals</h2></div>
        <div className="mt-5 grid gap-5 lg:grid-cols-[minmax(0,1fr)_420px]">
          <div>
            {goals.length === 0 ? (
              <div className="rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-[var(--muted)]">No goals yet. Create a target to begin monthly progress tracking.</div>
            ) : (
              <div className="grid gap-4 sm:grid-cols-2">
                {goals.map((goal) => <GoalCard key={goal.id} goal={goal} portfolioID={portfolio.id} snapshot={latestByGoal.get(goal.id) ?? null} />)}
              </div>
            )}
          </div>
          <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6">
            <p className="eyebrow">New goal</p><h3 className="mt-2 text-xl font-semibold">Set a target</h3>
            <CreateGoalForm portfolioID={portfolio.id} currency={portfolio.base_currency} />
          </aside>
        </div>
      </section>

      <section className="mt-12 border-t border-[var(--line)] pt-10" aria-labelledby="simulations-title">
        <div><p className="eyebrow">Stateless calculators</p><h2 className="mt-2 text-2xl font-semibold" id="simulations-title">Contribution scenarios</h2></div>
        <div className="mt-5 grid gap-6 xl:grid-cols-2">
          <SIPCalculator portfolioID={portfolio.id} currency={portfolio.base_currency} />
          <WhatIfCalculator portfolioID={portfolio.id} currency={portfolio.base_currency} />
        </div>
      </section>
    </main>
  );
}

function GoalCard({ goal, portfolioID, snapshot }: { goal: Goal; portfolioID: string; snapshot: MonthlyGoalSnapshot | null }) {
  const progress = snapshot ? Math.max(0, Math.min(100, Number(snapshot.progress_percentage))) : 0;
  return (
    <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6">
      <div className="flex items-start justify-between gap-4"><div><p className="text-xs font-bold uppercase tracking-[0.1em] text-[var(--brand)]">{goal.status}</p><h3 className="mt-2 text-xl font-semibold">{goal.name}</h3></div><span className="text-xs text-[var(--muted)]">Due {goal.target_date}</span></div>
      <p className="mt-5 text-2xl font-semibold">{money(goal.target_amount, goal.currency)}</p>
      {snapshot ? <>
        <div className="mt-5 h-2.5 overflow-hidden rounded-full bg-[#e4e7e1]"><div className="h-full rounded-full bg-[var(--brand)]" style={{ width: `${progress}%` }} /></div>
        <div className="mt-3 flex justify-between text-xs text-[var(--muted)]"><span>{formatPercent(snapshot.progress_percentage)} funded</span><span>{money(snapshot.remaining_amount, snapshot.currency)} remaining</span></div>
        <p className="mt-4 border-t border-[var(--line)] pt-4 text-sm text-[var(--muted)]">Required monthly contribution: <strong className="text-[var(--ink)]">{money(snapshot.required_monthly_contribution, snapshot.currency)}</strong></p>
      </> : <p className="mt-5 rounded-xl bg-[#f1f2ed] p-3 text-sm text-[var(--muted)]">Monthly progress appears after the month-end snapshot job runs.</p>}
      <GoalActions goal={goal} portfolioID={portfolioID} />
    </article>
  );
}

function money(value: string, currency: string) { return new Intl.NumberFormat("en-IN", { style: "currency", currency, maximumFractionDigits: 2 }).format(Number(value)); }
function formatPercent(value: string) { return `${new Intl.NumberFormat("en-IN", { maximumFractionDigits: 2 }).format(Number(value))}%`; }
