import type { CurrencyHealthScore } from "@/lib/types";

export function HealthScoreCard({ score }: { score: CurrencyHealthScore }) {
  const width = Math.max(0, Math.min(100, score.score));
  return (
    <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="eyebrow">{score.currency} health</p>
          <p className="mt-2 text-4xl font-semibold tracking-[-0.05em]">
            {score.score}<span className="text-xl text-[var(--muted)]">/{score.maximum}</span>
          </p>
        </div>
        <span className="rounded-full bg-[var(--brand-soft)] px-3 py-1 text-xs font-bold text-[var(--brand)]">Moderate</span>
      </div>
      <div className="mt-5 h-2.5 overflow-hidden rounded-full bg-[#e4e7e1]">
        <div className="h-full rounded-full bg-[var(--brand)]" style={{ width: `${width}%` }} />
      </div>
      <div className="mt-5 space-y-3">
        {score.components.map((component) => (
          <div className="flex items-center justify-between gap-4 text-sm" key={component.category}>
            <span className="capitalize text-[var(--muted)]">{component.category.replaceAll("_", " ")}</span>
            <span className="font-mono font-semibold">{component.points}/{component.maximum}</span>
          </div>
        ))}
      </div>
    </article>
  );
}
