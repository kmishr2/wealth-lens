"use client";

export default function AppError({ reset }: { reset: () => void }) {
  return (
    <main className="mx-auto grid min-h-[70vh] max-w-xl place-items-center px-6 text-center">
      <div>
        <p className="eyebrow">Connection problem</p>
        <h1 className="mt-3 text-4xl font-semibold tracking-[-0.04em]">
          The workspace could not load.
        </h1>
        <p className="mt-4 leading-7 text-[var(--muted)]">
          Check that the WealthLens backend and PostgreSQL are running, then try
          again.
        </p>
        <button
          className="focus-ring mt-7 rounded-xl bg-[var(--brand)] px-5 py-3 font-semibold text-white hover:bg-[var(--brand-strong)]"
          onClick={reset}
          type="button"
        >
          Try again
        </button>
      </div>
    </main>
  );
}
