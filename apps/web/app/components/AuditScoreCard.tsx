'use client';

import * as m from 'framer-motion/m';
import type { GateResponse } from '@/lib/types';
import { dimensionLabel, parseReasons, scoreTone } from '@/lib/reasons';
import { AuditScoreRing } from './AuditScoreRing';

interface AuditScoreCardProps {
  l1?: GateResponse;
  l2?: GateResponse;
}

const TONE_BG: Record<'good' | 'mid' | 'bad', string> = {
  good: 'bg-emerald-500',
  mid:  'bg-amber-500',
  bad:  'bg-red-500',
};
const TONE_TEXT: Record<'good' | 'mid' | 'bad', string> = {
  good: 'text-emerald-300',
  mid:  'text-amber-300',
  bad:  'text-red-300',
};

// Three-layer progressive disclosure:
//   Layer 1 — score rings (one per gate) — ALWAYS visible
//   Layer 2 — audit dimensions per gate    — collapsed; expand via <details>
//   Layer 3 — full rule / EEF / SPT chain  — collapsed; expand via inner <details>
export function AuditScoreCard({ l1, l2 }: AuditScoreCardProps) {
  if (!l1 && !l2) return null;

  return (
    <section className="rounded-xl border border-zinc-800 bg-zinc-900/40 p-5">
      <div className="mb-4 flex items-baseline justify-between gap-4">
        <h2 className="text-xs font-semibold uppercase tracking-wider text-zinc-300">
          GEM² Audit Score
        </h2>
        <span className="text-[11px] text-zinc-500">
          source: <code>gem2-tpmn-checker.fly.dev</code>
        </span>
      </div>

      {/* ── Layer 1: rings ─────────────────────────────────────────────── */}
      <div className="flex flex-wrap items-start justify-center gap-10 py-2">
        <AuditScoreRing
          score={l1?.score ?? -1}
          label="L1 P-check"
          verdict={l1?.verdict}
          size={140}
        />
        <AuditScoreRing
          score={l2 ? l2.score : -1}
          label="L2 O-check"
          verdict={l2?.verdict ?? 'skipped'}
          size={140}
          subtle={!l2}
        />
      </div>

      {/* ── Layer 2: dimensions (per gate, side by side) ───────────────── */}
      <details className="group mt-5 rounded-lg border border-zinc-800 bg-zinc-950/40 open:bg-zinc-950">
        <summary className="cursor-pointer select-none list-none px-4 py-2.5 text-sm text-zinc-300 hover:bg-zinc-900/60">
          <span className="inline-block w-4 text-zinc-500 transition-transform group-open:rotate-90">›</span>
          Show audit dimensions
        </summary>
        <div className="border-t border-zinc-800 p-4">
          <div className="grid gap-6 lg:grid-cols-2">
            <GateDimensions title="L1 P-check" gate={l1} />
            <GateDimensions title="L2 O-check" gate={l2} />
          </div>

          {/* ── Layer 3: full rule chain (nested deeper) ───────────────── */}
          <details className="group/inner mt-5 rounded-lg border border-zinc-800 bg-zinc-950/40 open:bg-zinc-950/60">
            <summary className="cursor-pointer select-none list-none px-4 py-2 text-xs text-zinc-400 hover:bg-zinc-900/60">
              <span className="inline-block w-4 text-zinc-500 transition-transform group-open/inner:rotate-90">›</span>
              Show all rules, SPT flags, EEF tags, and evidence references (technical)
            </summary>
            <div className="border-t border-zinc-800 p-4">
              <div className="grid gap-5 lg:grid-cols-2">
                <RawReasons title="L1 P-check raw reasons" gate={l1} />
                <RawReasons title="L2 O-check raw reasons" gate={l2} />
              </div>
            </div>
          </details>
        </div>
      </details>
    </section>
  );
}

// ── Per-gate dimension bars ──────────────────────────────────────────────

function GateDimensions({ title, gate }: { title: string; gate?: GateResponse }) {
  if (!gate) {
    return (
      <div>
        <h3 className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-zinc-500">
          {title} <span className="font-normal italic text-zinc-600">— skipped</span>
        </h3>
        <p className="text-xs text-zinc-600">
          L2 was not invoked because the L1 P-check already denied.
        </p>
      </div>
    );
  }
  const parsed = parseReasons(gate.reasons);
  if (parsed.dimensions.length === 0) {
    return (
      <div>
        <h3 className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
          {title}
        </h3>
        <p className="text-xs text-zinc-500">No dimension scores in response.</p>
      </div>
    );
  }
  return (
    <div>
      <h3 className="mb-2 flex items-baseline justify-between text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
        <span>{title}</span>
        <span className="font-mono text-zinc-500">composite {gate.score}/100</span>
      </h3>
      <ul className="space-y-2">
        {parsed.dimensions.map((d, idx) => {
          const tone = scoreTone(d.score);
          return (
            <li key={d.name}>
              <div className="flex items-baseline justify-between gap-2">
                <span className="text-sm text-zinc-200">{dimensionLabel(d.name)}</span>
                <span className={`font-mono text-xs ${TONE_TEXT[tone]}`}>{d.score}</span>
              </div>
              <div className="mt-1 h-1.5 w-full overflow-hidden rounded-full bg-zinc-900">
                <m.div
                  className={`h-full ${TONE_BG[tone]}`}
                  initial={{ width: '0%' }}
                  animate={{ width: `${d.score}%` }}
                  transition={{ duration: 0.7, ease: 'easeOut', delay: 0.05 * idx }}
                />
              </div>
              {d.note && (
                <p className="mt-1 text-[11px] leading-relaxed text-zinc-500">{d.note}</p>
              )}
            </li>
          );
        })}
      </ul>
    </div>
  );
}

// ── Raw reasons (Layer 3) ────────────────────────────────────────────────

function RawReasons({ title, gate }: { title: string; gate?: GateResponse }) {
  if (!gate) return null;
  const parsed = parseReasons(gate.reasons);
  return (
    <div>
      <h4 className="mb-2 text-[11px] font-semibold uppercase tracking-wider text-zinc-400">
        {title}
      </h4>

      {parsed.typeFinding && (
        <p className="mb-3 text-xs italic text-zinc-400">
          <Tag tag="TYPE" /> {parsed.typeFinding}
        </p>
      )}

      {parsed.rules.length > 0 && (
        <ul className="mb-3 space-y-1 text-xs">
          {parsed.rules.map((r) => (
            <li key={r.index} className="flex gap-2 font-mono text-zinc-300">
              <Tag tag={`RULE-${r.index}`} tone={r.verdict === 'PASS' ? 'good' : 'bad'} />
              <span className="flex-1">
                <span className="text-zinc-100">{r.rule}</span>
                <span className="text-zinc-500"> — </span>
                <span className={r.verdict === 'PASS' ? 'text-emerald-300' : 'text-red-300'}>
                  {r.verdict}
                </span>
                <span className="text-zinc-500">: {r.reason}</span>
              </span>
            </li>
          ))}
        </ul>
      )}

      {(parsed.sptCodes.length > 0 || parsed.eefFlags.length > 0) && (
        <div className="mb-3 space-y-1 text-xs">
          {parsed.sptCodes.map((c, i) => (
            <div key={`spt-${i}`} className="flex gap-2 font-mono text-rose-300">
              <Tag tag={`SPT-${c}`} tone="bad" />
              <span className="text-zinc-400">overclaim flagged</span>
            </div>
          ))}
          {parsed.eefFlags.map((f, i) => (
            <div key={`eef-${i}`} className="flex gap-2 font-mono text-amber-300">
              <Tag tag={`EEF-${f.tag}`} tone="mid" />
              <span className="text-zinc-400">{f.text}</span>
            </div>
          ))}
        </div>
      )}

      {parsed.evidence.length > 0 && (
        <ul className="space-y-1 text-xs">
          {parsed.evidence.map((e, i) => (
            <li key={`ev-${i}`} className="flex gap-2 font-mono text-emerald-300">
              <Tag tag={`EVIDENCE-${e.index ?? '⊥'}`} tone="good" />
              <span className="text-zinc-400">{e.excerpt}</span>
            </li>
          ))}
        </ul>
      )}

      <div className="mt-3 text-[10px] text-zinc-500">
        result_id: <code>{gate.meta?.result_id}</code> · {gate.meta?.duration_ms} ms
        {gate.meta?.usage?.estimated_cost_usd != null && (
          <> · ${gate.meta.usage.estimated_cost_usd.toFixed(4)}</>
        )}
      </div>
    </div>
  );
}

function Tag({ tag, tone = 'neutral' }: { tag: string; tone?: 'good' | 'mid' | 'bad' | 'neutral' }) {
  const cls =
    tone === 'good' ? 'bg-emerald-500/20 text-emerald-300'
    : tone === 'mid' ? 'bg-amber-500/20 text-amber-300'
    : tone === 'bad' ? 'bg-rose-500/20 text-rose-300'
    : 'bg-zinc-800 text-zinc-300';
  return (
    <span className={`shrink-0 rounded px-1.5 py-0.5 text-[10px] font-bold ${cls}`}>
      {tag}
    </span>
  );
}
