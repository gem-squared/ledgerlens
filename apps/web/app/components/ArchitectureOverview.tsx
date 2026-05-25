'use client';

import * as m from 'framer-motion/m';
import { fadeLeft, staggerContainer, SPRING_GENTLE } from '@/lib/motion';

const STEPS: Array<{ id: string; title: string; line: string }> = [
  { id: 'judge',     title: 'Judge',          line: 'Asks for trusted web data' },
  { id: 'buyer',     title: 'Buyer Agent',    line: 'Extracts intent and spend policy' },
  { id: 'brightdata',title: 'Bright Data',    line: 'Finds and fetches public evidence' },
  { id: 'seller',    title: 'Seller Offer',   line: 'Forms candidate data deal' },
  { id: 'gem2',      title: 'GEM\u00b2 Gate',      line: 'Checks if claims are grounded' },
  { id: 'x402',      title: 'x402 Settle',    line: 'Approves or blocks payment' },
  { id: 'bundle',    title: 'Audit Bundle',   line: 'Exports replayable record' },
];

export function ArchitectureOverview() {
  return (
    <m.section
      className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5"
      initial="hidden"
      whileInView="visible"
      viewport={{ once: true, amount: 0.3 }}
      variants={staggerContainer}
    >
      <div className="mb-4 flex items-baseline justify-between gap-4">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-300">
          Architecture Overview
        </h2>
        <span className="text-[11px] text-zinc-500">
          Verification infrastructure pipeline {'\u00b7'} one direction
        </span>
      </div>

      <m.div
        className="flex flex-col gap-3 lg:flex-row lg:items-stretch lg:gap-2"
        variants={staggerContainer}
      >
        {STEPS.map((s, i) => (
          <Step key={s.id} step={s} hasArrow={i < STEPS.length - 1} isGem2={s.id === 'gem2'} />
        ))}
      </m.div>

      <m.p
        className="mt-4 text-xs italic text-zinc-500"
        variants={fadeLeft}
        transition={{ duration: 0.4 }}
      >
        Every box has its own contract. Money only moves after the last green check.
      </m.p>
    </m.section>
  );
}

function Step({
  step,
  hasArrow,
  isGem2,
}: {
  step: { title: string; line: string };
  hasArrow: boolean;
  isGem2: boolean;
}) {
  return (
    <m.div
      className="flex flex-1 items-stretch gap-2 lg:flex-col lg:items-stretch"
      variants={fadeLeft}
      transition={SPRING_GENTLE}
    >
      <div className={`flex-1 rounded-lg border border-zinc-700 bg-zinc-950 p-3 transition hover:border-zinc-500${
        isGem2 ? ' glow-indigo' : ''
      }`}>
        <div className="text-xs font-semibold uppercase tracking-wider text-zinc-200">
          {step.title}
        </div>
        <p className="mt-1 text-[11px] leading-relaxed text-zinc-400">{step.line}</p>
      </div>
      {hasArrow && (
        <div className="flex items-center justify-center text-zinc-600 lg:py-0" aria-hidden="true">
          <span className="hidden text-lg lg:inline">{'\u2192'}</span>
          <span className="text-lg lg:hidden">{'\u2193'}</span>
        </div>
      )}
    </m.div>
  );
}
