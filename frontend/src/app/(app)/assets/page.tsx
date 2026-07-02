import type { Metadata } from "next";
import Link from "next/link";
import { redirect } from "next/navigation";
import { CreateAssetForm } from "@/components/create-asset-form";
import { apiRequest, ApiError } from "@/lib/api";
import { requireUser } from "@/lib/auth";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Asset } from "@/lib/types";

export const metadata: Metadata = { title: "Assets" };

async function listAssets() {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  try { return await apiRequest<Asset[]>("/assets?limit=100", { accessToken }); }
  catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) redirect("/auth/refresh?next=/assets");
      redirect("/login");
    }
    throw error;
  }
}

export default async function AssetsPage() {
  const [user, assets] = await Promise.all([requireUser(), listAssets()]);
  return (
    <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
      <div className="border-b border-[var(--line)] pb-9"><p className="eyebrow">Reference data</p><h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">Asset catalogue</h1><p className="mt-4 max-w-3xl leading-7 text-[var(--muted)]">Assets identify ledger positions and their valuation currency. They do not connect to a broker or authorize trading.</p></div>
      <div className="mt-9 grid gap-8 lg:grid-cols-[minmax(0,1fr)_420px]">
        <section aria-labelledby="asset-list-title">
          <div className="flex items-end justify-between"><div><p className="eyebrow">Available</p><h2 className="mt-2 text-2xl font-semibold" id="asset-list-title">Assets</h2></div><span className="text-sm text-[var(--muted)]">{assets.length} records</span></div>
          {assets.length === 0 ? <div className="mt-5 rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-[var(--muted)]">No assets have been created.</div> : (
            <div className="mt-5 grid gap-4 sm:grid-cols-2">
              {assets.map((asset) => (
                <Link className="focus-ring rounded-2xl border border-[var(--line)] bg-[var(--surface-strong)] p-5 transition hover:-translate-y-0.5 hover:shadow-[var(--shadow)]" href={`/assets/${asset.id}`} key={asset.id}>
                  <div className="flex items-start justify-between gap-3"><div><p className="text-lg font-bold">{asset.symbol}</p><p className="mt-1 line-clamp-2 text-sm text-[var(--muted)]">{asset.name}</p></div><span className="rounded-full border border-[var(--line)] px-2.5 py-1 text-xs font-bold text-[var(--muted)]">{asset.currency}</span></div>
                  <div className="mt-5 flex flex-wrap gap-2 border-t border-[var(--line)] pt-4 text-xs"><span className="rounded-full bg-[var(--brand-soft)] px-2.5 py-1 capitalize text-[var(--brand)]">{asset.asset_class.replaceAll("_", " ")}</span><span className="rounded-full bg-[#f1f2ed] px-2.5 py-1 text-[var(--muted)]">{asset.risk_category?.replaceAll("_", " ") ?? "Unclassified risk"}</span></div>
                </Link>
              ))}
            </div>
          )}
        </section>
        <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6 shadow-[var(--shadow)]"><p className="eyebrow">New reference</p><h2 className="mt-2 text-2xl font-semibold">Create asset</h2><p className="mt-3 text-sm leading-6 text-[var(--muted)]">Funds require an explicit equity, debt, or cash/other risk category for complete health scoring.</p><CreateAssetForm defaultCurrency={user.base_currency} /></aside>
      </div>
    </main>
  );
}
