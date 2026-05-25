import { HeroSection } from './components/HeroSection';
import { StickyNav } from './components/StickyNav';
import { ScrollSection } from './components/ScrollSection';
import { HowItWorks } from './components/HowItWorks';
import { DashboardShell } from './components/DashboardShell';
import { CaseRunner } from './components/CaseRunner';

export default function Home() {
  return (
    <main className="mx-auto max-w-6xl px-6">
      <StickyNav />

      <section id="hero" className="py-10">
        <HeroSection />
      </section>

      <ScrollSection id="how-it-works" speed={1.0}>
        <h2 className="section-heading">How It Works</h2>
        <p className="section-sub">
          Trust-gated verification in 7 steps — no grounded claim, no payment.
        </p>
        <HowItWorks />
      </ScrollSection>

      <ScrollSection id="evidence" speed={0.7}>
        <h2 className="section-heading">Real-Time Evidence &amp; Audit</h2>
        <p className="section-sub">
          Bright Data collects public-web proof. GEM² verifies every claim before payment.
        </p>
        <DashboardShell />
      </ScrollSection>

      <ScrollSection id="settlement" speed={1.3}>
        <h2 className="section-heading">Settlement Simulation</h2>
        <p className="section-sub">
          Deterministic replay — fixed BLOCKED / APPROVED outcomes for demo safety.
        </p>
        <CaseRunner />
      </ScrollSection>

      <footer className="py-12 border-t border-white/[0.08] text-xs text-zinc-500 text-center">
        Built for the Bright Data &quot;Web Data UNLOCKED&quot; Hackathon · LedgerLens by GEM².AI
      </footer>
    </main>
  );
}
