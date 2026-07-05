import type { PortfolioSnapshot } from "@/lib/types";

type Point = {
  date: string;
  fullyValued: boolean;
  value: number;
};

const width = 720;
const height = 220;
const padding = { top: 18, right: 18, bottom: 30, left: 18 };

export function SnapshotValueChart({ snapshots }: { snapshots: PortfolioSnapshot[] }) {
  const currencies = Array.from(
    new Set(
      snapshots.flatMap((snapshot) =>
        snapshot.total_values.map((total) => total.currency),
      ),
    ),
  ).sort();

  if (currencies.length === 0) {
    return (
      <p className="mt-5 rounded-2xl border border-dashed border-[#bdc6c0] bg-[var(--surface)] p-6 text-sm text-[var(--muted)]">
        Snapshot values are not available yet.
      </p>
    );
  }

  return (
    <div className="mt-5 grid gap-4 xl:grid-cols-2">
      {currencies.map((currency) => (
        <CurrencyChart
          currency={currency}
          key={currency}
          points={pointsForCurrency(snapshots, currency)}
        />
      ))}
    </div>
  );
}

function CurrencyChart({ currency, points }: { currency: string; points: Point[] }) {
  if (points.length === 0) return null;

  const values = points.map((point) => point.value);
  const minimum = Math.min(...values);
  const maximum = Math.max(...values);
  const spread = maximum - minimum;
  const plotWidth = width - padding.left - padding.right;
  const plotHeight = height - padding.top - padding.bottom;
  const coordinates = points.map((point, index) => ({
    ...point,
    x:
      points.length === 1
        ? padding.left + plotWidth / 2
        : padding.left + (index / (points.length - 1)) * plotWidth,
    y:
      spread === 0
        ? padding.top + plotHeight / 2
        : padding.top + ((maximum - point.value) / spread) * plotHeight,
  }));
  const polyline = coordinates.map((point) => `${point.x},${point.y}`).join(" ");
  const first = points[0];
  const last = points[points.length - 1];

  return (
    <article className="rounded-3xl border border-[var(--line)] bg-[var(--surface-strong)] p-5 sm:p-6">
      <div className="flex items-start justify-between gap-4">
        <div>
          <p className="eyebrow">{currency}</p>
          <h3 className="mt-2 text-xl font-semibold">Portfolio value</h3>
        </div>
        <div className="text-right">
          <p className="text-xs text-[var(--muted)]">Latest snapshot</p>
          <p className="mt-1 font-mono font-semibold">
            {formatMoney(last.value, currency)}
          </p>
        </div>
      </div>

      <svg
        aria-labelledby={`chart-title-${currency} chart-description-${currency}`}
        className="mt-5 h-auto w-full overflow-visible"
        role="img"
        viewBox={`0 0 ${width} ${height}`}
      >
        <title id={`chart-title-${currency}`}>{currency} portfolio value history</title>
        <desc id={`chart-description-${currency}`}>
          {points.length} daily snapshot values from {formatDate(first.date)} to {formatDate(last.date)}.
        </desc>
        {[0, 0.5, 1].map((ratio) => {
          const y = padding.top + ratio * plotHeight;
          return <line key={ratio} x1={padding.left} x2={width - padding.right} y1={y} y2={y} stroke="var(--line)" strokeDasharray="4 6" />;
        })}
        {points.length > 1 && (
          <polyline fill="none" points={polyline} stroke="var(--brand)" strokeLinecap="round" strokeLinejoin="round" strokeWidth="4" />
        )}
        {coordinates.map((point) => (
          <circle
            cx={point.x}
            cy={point.y}
            fill={point.fullyValued ? "var(--brand)" : "var(--surface-strong)"}
            key={`${currency}-${point.date}`}
            r={point.fullyValued ? 4 : 6}
            stroke={point.fullyValued ? "var(--brand)" : "var(--danger)"}
            strokeWidth={point.fullyValued ? 0 : 3}
          >
            <title>{`${formatDate(point.date)}: ${formatMoney(point.value, currency)}${point.fullyValued ? "" : " (incomplete valuation)"}`}</title>
          </circle>
        ))}
        <text fill="var(--muted)" fontSize="12" x={padding.left} y={height - 5}>{formatDate(first.date)}</text>
        <text fill="var(--muted)" fontSize="12" textAnchor="end" x={width - padding.right} y={height - 5}>{formatDate(last.date)}</text>
      </svg>

      <div className="mt-2 flex flex-wrap items-center justify-between gap-3 text-xs text-[var(--muted)]">
        <span>{points.length} daily snapshot{points.length === 1 ? "" : "s"}</span>
        <span className="inline-flex items-center gap-2">
          <span className="inline-block h-2.5 w-2.5 rounded-full border-2 border-[var(--danger)] bg-white" />
          Incomplete valuation
        </span>
      </div>
    </article>
  );
}

function pointsForCurrency(snapshots: PortfolioSnapshot[], currency: string) {
  return snapshots
    .flatMap((snapshot) => {
      const total = snapshot.total_values.find((item) => item.currency === currency);
      if (!total) return [];
      const value = Number(total.amount);
      if (!Number.isFinite(value)) return [];
      return [{ date: snapshot.snapshot_date, fullyValued: snapshot.is_fully_valued, value }];
    })
    .sort((left, right) => left.date.localeCompare(right.date));
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat("en-IN", {
    day: "2-digit",
    month: "short",
    year: "numeric",
    timeZone: "UTC",
  }).format(new Date(`${value}T00:00:00Z`));
}

function formatMoney(value: number, currency: string) {
  return new Intl.NumberFormat("en-IN", {
    style: "currency",
    currency,
    maximumFractionDigits: 2,
  }).format(value);
}
