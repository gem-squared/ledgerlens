'use client';

import { motion } from 'framer-motion';
import { SPRING_GENTLE, fadeUp, staggerContainer } from '@/lib/motion';

export function HeroSection() {
  return (
    <motion.header
      className="mb-8"
      initial="hidden"
      animate="visible"
      variants={staggerContainer}
    >
      <motion.div
        className="flex flex-wrap items-baseline justify-between gap-4"
        variants={fadeUp}
        transition={{ duration: 0.5 }}
      >
        <div>
          <span className="inline-block rounded-full bg-simBadge px-3 py-1 text-xs font-semibold uppercase tracking-wider text-white">
            Simulation Mode
          </span>
          <h1 className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-4xl font-bold tracking-tight">
            <span>LedgerLens</span>
            <span aria-hidden className="text-3xl font-light text-zinc-500">×</span>
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src="/brightdata-logo.png"
              alt="Bright Data"
              className="h-7 w-auto sm:h-8"
            />
          </h1>
          <p className="mt-1 text-lg text-zinc-400">No grounded claim, no payment.</p>
        </div>
        <a
          href="https://github.com/gem-squared/ledgerlens"
          className="rounded-md border border-zinc-700 px-3 py-1.5 text-xs text-zinc-400 transition hover:border-zinc-500 hover:text-zinc-200"
        >
          source ↗
        </a>
      </motion.div>

      <motion.p
        className="mt-4 max-w-3xl text-sm text-zinc-400"
        variants={fadeUp}
        transition={{ duration: 0.4 }}
      >
        LedgerLens is a trust-gated commerce layer for autonomous agents. A buyer agent
        interprets the judge&apos;s request, Bright Data collects public evidence, and GEM²
        audits the seller&apos;s claim before any payment is allowed. Settlement is x402-shaped
        simulation: the trust gate is real.
      </motion.p>

      <motion.p
        className="mt-3 max-w-3xl text-sm italic text-zinc-300"
        variants={fadeUp}
        transition={{ duration: 0.4 }}
      >
        Fast agents are dangerous if they spend before verification. LedgerLens deliberately waits.
      </motion.p>

      {/* Three product pillars */}
      <motion.div
        className="mt-5 grid gap-3 sm:grid-cols-3"
        variants={staggerContainer}
      >
        <PillarCard title="Bright Data Evidence">
          SERP + Unlocker + Browser + MCP collect public-web proof.
        </PillarCard>
        <PillarCard title="GEM² Trust Gate">
          L1 P-check + L2 O-check classify claims · ⊢ ⊨ ⊬ ⊥
        </PillarCard>
        <PillarCard title="x402 Settlement">
          L3 releases or blocks simulated payment.
        </PillarCard>
      </motion.div>

      {/* Scroll indicator */}
      <motion.div
        className="mt-10 flex justify-center"
        variants={fadeUp}
        transition={{ delay: 0.8 }}
      >
        <button
          type="button"
          onClick={() => document.getElementById('how-it-works')?.scrollIntoView({ behavior: 'smooth' })}
          className="text-zinc-500 hover:text-zinc-300 transition-colors animate-bounce"
          aria-label="Scroll to content"
        >
          <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2">
            <path d="M7 13l5 5 5-5M7 6l5 5 5-5" />
          </svg>
        </button>
      </motion.div>
    </motion.header>
  );
}

function PillarCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <motion.div
      className="glass-panel"
      variants={fadeUp}
      transition={SPRING_GENTLE}
    >
      <div className="text-xs font-semibold uppercase tracking-wider text-zinc-500">
        {title}
      </div>
      <p className="mt-1 text-sm leading-relaxed text-zinc-300">{children}</p>
    </motion.div>
  );
}
