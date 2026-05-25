import { HeroSection } from './components/HeroSection';
import { ArchitectureOverview } from './components/ArchitectureOverview';
import { DashboardShell } from './components/DashboardShell';

export default function Home() {
  return (
    <main className="mx-auto max-w-6xl px-6 py-10">
      <HeroSection />

      <div className="mt-6">
        <DashboardShell />
      </div>

      {/* Architecture overview moved below the data so the dashboard leads with proof.
          The diagram is supporting context for visitors who want to know HOW the
          numbers above are produced. */}
      <div className="mt-14 border-t border-zinc-800 pt-8">
        <ArchitectureOverview />
      </div>

      <footer className="mt-12 border-t border-zinc-800 pt-6 text-xs text-zinc-500">
        Built for the Bright Data &quot;Web Data UNLOCKED&quot; Hackathon. Three modes:{' '}
        <strong>LIVE</strong> 20–45s · <strong>PRE-WARMED</strong> 5–8s ·{' '}
        <strong>REPLAY</strong> instant. L1 / L2 audit by{' '}
        <code>gem2-tpmn-checker.fly.dev</code>.
      </footer>
    </main>
  );
}
