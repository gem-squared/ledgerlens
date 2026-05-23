// LedgerLens — demo UI placeholder (Unit 1 scaffold).
// Unit 5 replaces this with the real buyer/seller/audit panel + LIVE/REPLAY toggle.

export default function Home() {
  return (
    <main className="mx-auto max-w-4xl px-6 py-16">
      <span className="inline-block rounded-full bg-simBadge px-3 py-1 text-xs font-semibold uppercase tracking-wider text-white">
        Simulation Mode
      </span>

      <h1 className="mt-6 text-4xl font-bold tracking-tight">LedgerLens</h1>

      <p className="mt-4 text-xl text-zinc-700 dark:text-zinc-300">
        No grounded claim, no payment.
      </p>

      <p className="mt-6 leading-relaxed text-zinc-600 dark:text-zinc-400">
        x402-native Agent-to-Agent Payments with a Bright Data web-evidence layer and
        a GEM² Trust Gate before settlement. We simulate settlement, but the trust
        gate is real.
      </p>

      <div className="mt-10 rounded-lg border border-zinc-200 bg-white p-6 shadow-sm dark:border-zinc-800 dark:bg-zinc-900">
        <p className="text-sm font-semibold uppercase tracking-wider text-zinc-500">
          Unit 1 scaffold
        </p>
        <p className="mt-2 text-zinc-700 dark:text-zinc-300">
          Backend API endpoints are stubbed to <code>501 Not Implemented</code> until
          Units 2–4 land. The buyer / seller / audit panel UI is the deliverable of
          Unit 5. This page exists so the Next.js + Tailwind toolchain is verified
          green at T-2.
        </p>
      </div>
    </main>
  );
}
