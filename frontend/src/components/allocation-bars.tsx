import type { AssetClassAllocation } from "@/lib/types";

const colors = ["#1f5a44", "#c79038", "#58766a", "#8c6840", "#80958b"];

export function AllocationBars({ items }: { items: AssetClassAllocation[] }) {
  if (items.length === 0) {
    return <p className="text-sm text-[var(--muted)]">No valued allocation yet.</p>;
  }

  return (
    <div className="space-y-4">
      {items.map((item, index) => {
        const percentage = Number(item.percentage);
        const width = Number.isFinite(percentage)
          ? Math.max(0, Math.min(100, percentage))
          : 0;
        return (
          <div key={`${item.currency}-${item.asset_class}`}>
            <div className="mb-1.5 flex items-center justify-between gap-4 text-sm">
              <span className="font-semibold capitalize">
                {item.asset_class.replaceAll("_", " ")}
                <span className="ml-2 text-xs font-normal text-[var(--muted)]">{item.currency}</span>
              </span>
              <span className="font-mono text-xs font-semibold">{formatPercentage(item.percentage)}</span>
            </div>
            <div className="h-2.5 overflow-hidden rounded-full bg-[#e4e7e1]">
              <div
                className="h-full rounded-full"
                style={{ backgroundColor: colors[index % colors.length], width: `${width}%` }}
              />
            </div>
          </div>
        );
      })}
    </div>
  );
}

function formatPercentage(value: string) {
  return `${new Intl.NumberFormat("en-IN", { maximumFractionDigits: 2 }).format(Number(value))}%`;
}
