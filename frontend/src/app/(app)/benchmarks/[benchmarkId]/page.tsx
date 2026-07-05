import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { CreateBenchmarkObservationForm } from "@/components/create-benchmark-observation-form";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Benchmark, BenchmarkObservation } from "@/lib/types";

export const metadata: Metadata = { title: "Benchmark observations" };

async function loadBenchmark(benchmarkID: string) {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  const encoded = encodeURIComponent(benchmarkID);
  try {
    const benchmarks = await apiRequest<Benchmark[]>("/benchmarks?limit=100", { accessToken });
    const benchmark = benchmarks.find((item) => item.id === benchmarkID);
    if (!benchmark) notFound();
    const observations = await apiRequest<BenchmarkObservation[]>(`/benchmarks/${encoded}/observations?limit=100`, { accessToken });
    return { benchmark, observations };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) redirect(`/auth/refresh?next=${encodeURIComponent(`/benchmarks/${benchmarkID}`)}`);
      redirect("/login");
    }
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
}

export default async function BenchmarkPage({ params }: { params: Promise<{ benchmarkId: string }> }) {
  const { benchmarkId } = await params;
  const { benchmark, observations } = await loadBenchmark(benchmarkId);
  return <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
    <Link className="focus-ring text-sm font-semibold text-[var(--brand)] hover:underline" href="/benchmarks">← Benchmarks</Link>
    <div className="mt-7 border-b border-[var(--line)] pb-9"><div className="flex items-center gap-3"><p className="eyebrow">{benchmark.source}</p><span className="rounded-full border border-[var(--line)] bg-white px-3 py-1 text-xs font-bold text-[var(--muted)]">{benchmark.currency}</span></div><h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">{benchmark.code}</h1><p className="mt-3 text-[var(--muted)]">{benchmark.name}</p></div>
    <div className="mt-9 grid gap-8 lg:grid-cols-[minmax(0,1fr)_380px]">
      <section><div className="flex items-end justify-between"><div><p className="eyebrow">Exact dates</p><h2 className="mt-2 text-2xl font-semibold">Observations</h2></div><span className="text-sm text-[var(--muted)]">{observations.length} records</span></div>
        {observations.length === 0 ? <div className="mt-5 rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-[var(--muted)]">No observations recorded.</div> : <div className="mt-5 overflow-hidden rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)]"><table className="w-full text-left text-sm"><thead className="bg-[#f1f2ed] text-xs uppercase tracking-wide text-[var(--muted)]"><tr><th className="px-5 py-3">Date</th><th className="px-5 py-3">Value</th><th className="px-5 py-3">Source</th></tr></thead><tbody className="divide-y divide-[var(--line)]">{observations.map((observation) => <tr key={observation.id}><td className="px-5 py-4 font-semibold">{observation.observation_date}</td><td className="px-5 py-4 font-mono">{number(observation.value)}</td><td className="px-5 py-4 text-[var(--muted)]">{observation.source}</td></tr>)}</tbody></table></div>}
      </section>
      <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6 shadow-[var(--shadow)]"><p className="eyebrow">Manual data</p><h2 className="mt-2 text-2xl font-semibold">Record observation</h2><p className="mt-3 text-sm leading-6 text-[var(--muted)]">Use dates that exactly match portfolio snapshots for comparison.</p><CreateBenchmarkObservationForm benchmarkID={benchmark.id} /></aside>
    </div>
  </main>;
}

function number(value: string) { return new Intl.NumberFormat("en-IN", { maximumFractionDigits: 4 }).format(Number(value)); }
