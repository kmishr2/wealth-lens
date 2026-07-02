import type { Metadata } from "next";
import Link from "next/link";
import { notFound, redirect } from "next/navigation";
import { CreatePriceForm } from "@/components/create-price-form";
import { apiRequest, ApiError } from "@/lib/api";
import { getAccessToken, getRefreshToken } from "@/lib/session";
import type { Asset, AssetPrice } from "@/lib/types";

export const metadata: Metadata = { title: "Asset prices" };

async function loadAsset(assetID: string) {
  const accessToken = await getAccessToken();
  if (!accessToken) redirect("/login");
  const encoded = encodeURIComponent(assetID);
  try {
    const [asset, prices] = await Promise.all([
      apiRequest<Asset>(`/assets/${encoded}`, { accessToken }),
      apiRequest<AssetPrice[]>(`/assets/${encoded}/prices?limit=100`, { accessToken }),
    ]);
    return { asset, prices };
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) {
      if (await getRefreshToken()) redirect(`/auth/refresh?next=${encodeURIComponent(`/assets/${assetID}`)}`);
      redirect("/login");
    }
    if (error instanceof ApiError && error.status === 404) notFound();
    throw error;
  }
}

export default async function AssetPage({ params }: { params: Promise<{ assetId: string }> }) {
  const { assetId } = await params;
  const { asset, prices } = await loadAsset(assetId);
  return (
    <main className="mx-auto max-w-7xl px-6 py-10 lg:px-10 lg:py-14">
      <Link className="focus-ring text-sm font-semibold text-[var(--brand)] hover:underline" href="/assets">← Asset catalogue</Link>
      <div className="mt-7 border-b border-[var(--line)] pb-9"><div className="flex items-center gap-3"><p className="eyebrow">{asset.asset_class.replaceAll("_", " ")}</p><span className="rounded-full border border-[var(--line)] bg-white px-3 py-1 text-xs font-bold text-[var(--muted)]">{asset.currency}</span></div><h1 className="mt-3 text-4xl font-semibold tracking-[-0.045em] sm:text-5xl">{asset.symbol}</h1><p className="mt-3 text-[var(--muted)]">{asset.name}{asset.exchange ? ` · ${asset.exchange}` : ""}</p></div>
      <div className="mt-9 grid gap-8 lg:grid-cols-[minmax(0,1fr)_380px]">
        <section><div className="flex items-end justify-between"><div><p className="eyebrow">Immutable observations</p><h2 className="mt-2 text-2xl font-semibold">Price history</h2></div><span className="text-sm text-[var(--muted)]">{prices.length} records</span></div>
          {prices.length === 0 ? <div className="mt-5 rounded-3xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-8 text-[var(--muted)]">No explicit prices recorded.</div> : <div className="mt-5 overflow-hidden rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)]"><table className="w-full text-left text-sm"><thead className="bg-[#f1f2ed] text-xs uppercase tracking-wide text-[var(--muted)]"><tr><th className="px-5 py-3">Price</th><th className="px-5 py-3">Observed</th><th className="px-5 py-3">Source</th></tr></thead><tbody className="divide-y divide-[var(--line)]">{prices.map((price) => <tr key={price.id}><td className="px-5 py-4 font-mono font-semibold">{money(price.price, price.currency)}</td><td className="px-5 py-4 text-[var(--muted)]">{new Intl.DateTimeFormat("en-IN", { dateStyle: "medium", timeStyle: "short" }).format(new Date(price.priced_at))}</td><td className="px-5 py-4">{price.source}</td></tr>)}</tbody></table></div>}
        </section>
        <aside className="h-fit rounded-3xl border border-[var(--line)] bg-[var(--surface)] p-6 shadow-[var(--shadow)]"><p className="eyebrow">Manual observation</p><h2 className="mt-2 text-2xl font-semibold">Record price</h2><p className="mt-3 text-sm leading-6 text-[var(--muted)]">Historical prices are append-only. Automated provider ingestion remains separate.</p><CreatePriceForm assetID={asset.id} currency={asset.currency} /></aside>
      </div>
    </main>
  );
}

function money(value: string, currency: string) { return new Intl.NumberFormat("en-IN", { style: "currency", currency, maximumFractionDigits: 4 }).format(Number(value)); }
