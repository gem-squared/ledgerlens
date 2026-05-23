import { CaseRunner } from './components/CaseRunner';

export default function Home() {
  return (
    <main className="mx-auto max-w-6xl px-6 py-10">
      <header className="mb-8 flex flex-wrap items-end justify-between gap-4">
        <div>
          <span className="inline-block rounded-full bg-simBadge px-3 py-1 text-xs font-semibold uppercase tracking-wider text-white">
            Simulation Mode
          </span>
          <h1 className="mt-3 text-4xl font-bold tracking-tight">LedgerLens</h1>
          <p className="mt-1 text-lg text-zinc-400">No grounded claim, no payment.</p>
          <p className="mt-3 max-w-3xl text-sm text-zinc-400">
            x402-native Agent-to-Agent Payments with a Bright Data web-evidence layer and a GEM² Trust Gate before settlement.{' '}
            <span className="italic">We simulate settlement, but the trust gate is real.</span>
          </p>
        </div>
        <a
          href="https://github.com/gem-squared/ledgerlens"
          className="rounded-md border border-zinc-700 px-3 py-1.5 text-xs text-zinc-400 transition hover:border-zinc-500 hover:text-zinc-200"
        >
          source ↗
        </a>
      </header>

      <CaseRunner />

      <footer className="mt-12 border-t border-zinc-800 pt-6 text-xs text-zinc-500">
        Built for the Bright Data &quot;Web Data UNLOCKED&quot; Hackathon. Backend on{' '}
        <code>127.0.0.1:8082</code> via Next.js rewrite at <code>/api/*</code>. L1 / L2 audit by{' '}
        <code>gem2-tpmn-checker.fly.dev</code>.
      </footer>
    </main>
  );
}
