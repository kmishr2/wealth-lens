import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";
import { CreateBenchmarkForm } from "@/components/create-benchmark-form";
import { apiRequest, ApiError } from "@/lib/api";
import { requireUser } from "@/lib/auth";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Benchmark } from "@/lib/types";

export const metadata: Metadata = { title: "Benchmarks" };

async function listBenchmarks() {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  try { return await apiRequest<Benchmark[]>("/benchmarks?limit=100", { accessToken }); }
  catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) redirect("/auth/refresh?next=/benchmarks");
      redirect("/login");
    }
    throw error;
  }
}

export default async function BenchmarksPage() {
  const [user, benchmarks] = await Promise.all([requireUser(), listBenchmarks()]);
  return <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
    <div className="border-b border-[var(--line)] pb-9"><p className="eyebrow">Reference series</p><h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">Benchmarks</h1><p className="mt-4 max-w-3xl leading-7 text-[var(--muted)]">Benchmark assumptions are never hardcoded. Create an explicit series and record immutable dated observations before comparing it with a portfolio.</p></div>
    <div className="mt-9 grid gap-8 lg:grid-cols-[minmax(0,1fr)_420px]">
      <section><div className="flex items-end justify-between"><div><p className="eyebrow">Configured</p><h2 className="mt-2 text-2xl font-semibold">Reference series</h2></div><span className="text-sm text-[var(--muted)]">{benchmarks.length} records</span></div>
        {benchmarks.length === 0 ? <div className="mt-5 rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-[var(--muted)]">No benchmarks configured.</div> : <div className="mt-5 grid gap-4 sm:grid-cols-2">{benchmarks.map((benchmark) => <Link className="focus-ring rounded-2xl border border-[var(--line)] bg-[var(--surface-strong)] p-5 transition hover:-translate-y-0.5 hover:shadow-[var(--shadow)]" href={`/benchmarks/${benchmark.id}`} key={benchmark.id}><div className="flex items-start justify-between gap-3"><div><p className="text-lg font-bold">{benchmark.code}</p><p className="mt-1 text-sm text-[var(--muted)]">{benchmark.name}</p></div><span className="rounded-full border border-[var(--line)] px-2.5 py-1 text-xs font-bold text-[var(--muted)]">{benchmark.currency}</span></div><p className="mt-5 border-t border-[var(--line)] pt-4 text-xs text-[var(--muted)]">Source: {benchmark.source}</p></Link>)}</div>}
      </section>
      <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6 shadow-[var(--shadow)]"><p className="eyebrow">New series</p><h2 className="mt-2 text-2xl font-semibold">Create benchmark</h2><CreateBenchmarkForm defaultCurrency={user.base_currency} /></aside>
    </div>
  </main>;
}
