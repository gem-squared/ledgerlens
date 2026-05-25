'use client';

import { useEffect, useRef, useState } from 'react';
import type { StepEvent, StepStatus, RunMode } from '@/lib/types';
import * as m from 'framer-motion/m';
import { SPRING_SNAPPY } from '@/lib/motion';

// ── Step catalog ──────────────────────────────────────────────────────────
// id MUST match the backend's StepEvent.step exactly (see internal/api/deals_stream.go).
// `label` is the static panel name. The dynamic action verb comes from the
// SSE event's own `label` field when an event for this step has arrived.
export const STEP_CATALOG: Array<{
  id: string;
  title: string;
  hint: string;
}> = [
  { id: 'judge_request',     title: 'Judge Request',           hint: 'Receiving your data-purchase request…' },
  { id: 'buyer_intent',      title: 'Buyer Agent',             hint: 'Interpreting need and policy from the request…' },
  { id: 'brightdata_search', title: 'Bright Data Search',      hint: 'Searching the public web for candidate sources…' },
  { id: 'brightdata_fetch',  title: 'Bright Data Evidence',    hint: 'Fetching evidence receipts via Web Unlocker…' },
  { id: 'seller_offer',      title: 'Seller Offer Agent',      hint: 'Constructing a candidate deal from the evidence…' },
  { id: 'l1',                title: 'GEM² L1 P-check',         hint: 'Auditing whether the seller claim is grounded…' },
  { id: 'l2',                title: 'GEM² L2 O-check',         hint: 'Verifying decision packet postconditions…' },
  { id: 'l3',                title: 'L3 Trust Gate',           hint: 'Composing the final approve/block verdict…' },
  { id: 'settle',            title: 'x402 Settlement',         hint: 'Either issuing a simulated receipt or holding the wallet…' },
  { id: 'final_report',      title: 'Final Report',            hint: 'Composing the judge-readable summary…' },
];

// Expected per-step durations (ms). Drives the synthetic-progress curve so
// the bar advances smoothly even between actual SSE events. Calibrated from
// observed live runs (~30s total).
const STEP_DURATIONS_LIVE: Record<string, number> = {
  judge_request:     150,
  buyer_intent:      3000,
  brightdata_search: 4000,
  brightdata_fetch:  4000,
  seller_offer:      3000,
  l1:                8000,
  l2:                8000,
  l3:                400,
  settle:            300,
  final_report:      300,
};
const STEP_DURATIONS_PREWARMED: Record<string, number> = {
  judge_request:     50,
  buyer_intent:      1000,
  brightdata_search: 100,
  brightdata_fetch:  100,
  seller_offer:      800,
  l1:                2000,
  l2:                2000,
  l3:                100,
  settle:            100,
  final_report:      100,
};
function durationsFor(mode: RunMode): Record<string, number> {
  return mode === 'prewarmed' ? STEP_DURATIONS_PREWARMED : STEP_DURATIONS_LIVE;
}

// Rotating microcopy when no SSE event has arrived for >3s.
const HEARTBEAT_MICROCOPY = [
  'Searching the public web with Bright Data…',
  'Fetching evidence receipts…',
  'Auditing seller claim with GEM²…',
  'Checking payment policy…',
  'Preparing final report…',
  'Fast agents are dangerous if they spend before verification. LedgerLens deliberately waits.',
];

// Status that ends an active step (UI snaps to 100% for that step).
const TERMINAL_STATUSES: ReadonlySet<StepStatus> = new Set([
  'passed', 'blocked', 'failed', 'skipped', 'settled', 'rejected', 'escalated',
]);

// Status that should halt all animation (run done, blocked, or errored).
function isRunHaltingStatus(stepId: string, status: StepStatus): boolean {
  // L1/L2 blocked → still continue to L3/settle (which will also block).
  // settle terminal status → run done.
  // final_report passed → run done.
  if (stepId === 'final_report' && TERMINAL_STATUSES.has(status)) return true;
  if (stepId === 'settle' && (status === 'blocked' || status === 'failed')) {
    // settlement blocked path — final_report may still emit afterwards, so don't halt yet
    return false;
  }
  return false;
}

export interface AgentFlowTimelineProps {
  /** which run mode is in flight — drives synthetic step-duration table. */
  mode: RunMode;
  /** when the run actually started (Date.now()). null = not running yet. */
  runStartedAt: number | null;
  /** SSE events received so far, in arrival order. */
  events: StepEvent[];
  /** true once the result event arrives or an error fires (animation halts). */
  done: boolean;
  /** total events without one for > 3s triggers heartbeat microcopy. */
  lastEventAt: number | null;
}

interface StepState {
  status: StepStatus;
  label: string;            // dynamic label from latest event (fallback to hint)
  startedAt: number | null; // when 'running' arrived
  endedAt: number | null;   // when terminal arrived
}

export function AgentFlowTimeline(props: AgentFlowTimelineProps) {
  const { mode, runStartedAt, events, done, lastEventAt } = props;
  const durations = durationsFor(mode);
  const totalExpected = Object.values(durations).reduce((a, b) => a + b, 0);

  // ── Aggregate SSE events into per-step state ────────────────────────────
  const stepStates: Record<string, StepState> = {};
  for (const s of STEP_CATALOG) {
    stepStates[s.id] = { status: 'idle', label: s.hint, startedAt: null, endedAt: null };
  }
  for (const e of events) {
    const cur = stepStates[e.step] || { status: 'idle', label: e.label, startedAt: null, endedAt: null };
    if (e.status === 'running' && cur.startedAt === null) {
      cur.startedAt = Date.parse(e.ts) || Date.now();
    }
    if (TERMINAL_STATUSES.has(e.status)) {
      cur.endedAt = Date.parse(e.ts) || Date.now();
    }
    cur.status = e.status;
    cur.label = e.label || cur.label;
    stepStates[e.step] = cur;
  }
  // Currently-running step (last one without a terminal event)
  const currentStepId = STEP_CATALOG.find((s) => {
    const st = stepStates[s.id];
    return st.status === 'running' || (st.status !== 'idle' && !TERMINAL_STATUSES.has(st.status));
  })?.id ?? null;

  // ── Continuous animation tick (requestAnimationFrame) ───────────────────
  const [tick, setTick] = useState(0);
  const rafRef = useRef<number | null>(null);

  useEffect(() => {
    if (!runStartedAt || done) {
      if (rafRef.current !== null) cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
      return;
    }
    let alive = true;
    const loop = () => {
      if (!alive) return;
      setTick((t) => (t + 1) % 1_000_000);
      rafRef.current = requestAnimationFrame(loop);
    };
    rafRef.current = requestAnimationFrame(loop);
    return () => {
      alive = false;
      if (rafRef.current !== null) cancelAnimationFrame(rafRef.current);
      rafRef.current = null;
    };
  }, [runStartedAt, done]);

  const now = Date.now();
  void tick; // tick is the heartbeat — recompute below

  // Compute global synthetic progress (capped at 0.95 until 'done')
  const elapsed = runStartedAt ? now - runStartedAt : 0;
  let globalSynthetic = runStartedAt ? Math.min(0.95, elapsed / totalExpected) : 0;
  if (done) globalSynthetic = 1;

  // Compute per-step progress on the active step
  let activeStepFrac = 0;
  if (currentStepId) {
    const st = stepStates[currentStepId];
    const startedAt = st.startedAt ?? runStartedAt ?? now;
    const stepEl = now - startedAt;
    const stepDur = durations[currentStepId] || 3000;
    activeStepFrac = Math.min(0.95, stepEl / stepDur);
  }

  // ── Heartbeat / "still verifying…" microcopy ────────────────────────────
  const sinceLast = lastEventAt ? now - lastEventAt : 0;
  const showHeartbeat = !done && runStartedAt && sinceLast > 3000;
  const microcopyIdx =
    Math.floor((now - (runStartedAt ?? now)) / 4000) % HEARTBEAT_MICROCOPY.length;

  // Final verdict from settle/final_report
  const settleStatus = stepStates['settle']?.status;
  const isBlocked =
    settleStatus === 'blocked' || stepStates['l3']?.status === 'blocked' ||
    stepStates['final_report']?.status === 'blocked';
  const isApproved = settleStatus === 'settled';

  return (
    <div className="space-y-5">
      {/* ── Global progress bar + elapsed time ──────────────────────────── */}
      <div className="space-y-2">
        <div className="flex items-baseline justify-between">
          <div className="text-xs font-semibold uppercase tracking-wider text-zinc-400">
            Verification progress
          </div>
          <div className="text-xs text-zinc-500">
            {runStartedAt ? (
              <>
                <span className="font-mono text-zinc-300">{(elapsed / 1000).toFixed(1)}s</span>
                {' / target ~'}
                <span className="font-mono">{(totalExpected / 1000).toFixed(0)}s</span>
                {' '}({mode.toUpperCase()})
              </>
            ) : (
              <>idle</>
            )}
          </div>
        </div>
        <div className="h-2 w-full overflow-hidden rounded-full bg-zinc-900">
          <div
            className={done && isBlocked
              ? 'h-full bg-red-500 transition-all duration-200'
              : done
                ? 'h-full bg-emerald-500 transition-all duration-200'
                : 'll-bar-active h-full transition-all duration-200'}
            style={{ width: `${(globalSynthetic * 100).toFixed(2)}%` }}
          />
        </div>
        <div
          className="mt-1 h-1 w-full rounded-full glow-indigo transition-opacity duration-300"
          style={{ opacity: done ? 0 : globalSynthetic * 0.6 }}
        />
        {showHeartbeat && (
          <div className="flex items-center gap-2 text-xs text-zinc-400">
            <span className="ll-active inline-block h-2 w-2 rounded-full bg-simBadge" />
            <span className="italic">{HEARTBEAT_MICROCOPY[microcopyIdx]}</span>
          </div>
        )}
      </div>

      {/* ── Step list ───────────────────────────────────────────────────── */}
      <ol className="space-y-2">
        {STEP_CATALOG.map((s, idx) => {
          const st = stepStates[s.id];
          const isActive = s.id === currentStepId && !done;
          const isTerminal = TERMINAL_STATUSES.has(st.status);
          return (
            <li key={s.id}>
              <StepCard
                index={idx + 1}
                title={s.title}
                label={st.label}
                status={st.status}
                isActive={isActive}
                activeFrac={isActive ? activeStepFrac : (isTerminal ? 1 : 0)}
              />
            </li>
          );
        })}
      </ol>

      {/* ── Mode-aware bottom strip (only while running) ────────────────── */}
      {runStartedAt && !done && (
        <p className="text-center text-xs italic text-zinc-500">
          Fast agents are dangerous if they spend before verification. LedgerLens deliberately waits.
        </p>
      )}
      {done && isApproved && (
        <p className="text-center text-sm font-medium text-emerald-300">
          ✓ Verification complete. Simulated settlement issued.
        </p>
      )}
      {done && isBlocked && (
        <p className="text-center text-sm font-medium text-red-300">
          ✕ Verification complete. Payment blocked — a fast agent would have paid. LedgerLens waited.
        </p>
      )}
    </div>
  );
}

// ── Per-step card ─────────────────────────────────────────────────────────

interface StepCardProps {
  index: number;
  title: string;
  label: string;
  status: StepStatus;
  isActive: boolean;
  activeFrac: number; // 0..1
}

function StepCard({ index, title, label, status, isActive, activeFrac }: StepCardProps) {
  let bg = 'bg-zinc-900/40 border-zinc-800';
  let glow = '';
  let iconBg = 'bg-zinc-800 text-zinc-500';
  let iconChar = String(index);
  let labelClass = 'text-zinc-500';

  switch (status) {
    case 'running':
      bg = 'bg-indigo-500/5 border-indigo-500/40';
      glow = 'glow-indigo';
      iconBg = 'bg-indigo-500 text-white';
      iconChar = '▸';
      labelClass = 'text-zinc-200';
      break;
    case 'passed':
    case 'settled':
      bg = 'bg-emerald-500/5 border-emerald-500/30';
      glow = 'glow-green';
      iconBg = 'bg-emerald-500 text-white';
      iconChar = '✓';
      labelClass = 'text-zinc-300';
      break;
    case 'blocked':
    case 'failed':
      bg = 'bg-red-500/10 border-red-500/40';
      glow = 'glow-red';
      iconBg = 'bg-red-500 text-white';
      iconChar = '✕';
      labelClass = 'text-red-200';
      break;
    case 'skipped':
      bg = 'bg-zinc-900/40 border-zinc-800';
      iconBg = 'bg-zinc-800 text-zinc-500';
      iconChar = '—';
      labelClass = 'text-zinc-500 italic';
      break;
    case 'rejected':
      bg = 'bg-amber-500/10 border-amber-500/40';
      iconBg = 'bg-amber-500 text-white';
      iconChar = '?';
      labelClass = 'text-amber-200';
      break;
    case 'escalated':
      bg = 'bg-amber-500/10 border-amber-500/40';
      iconBg = 'bg-amber-500 text-white';
      iconChar = '↑';
      labelClass = 'text-amber-200';
      break;
    case 'idle':
    default:
      break;
  }

  return (
    <m.div
      className={`relative overflow-hidden rounded-lg border ${bg} ${glow} px-4 py-3`}
      animate={{
        scale: isActive ? 1.02 : 1,
        x: status === 'blocked' || status === 'failed' ? [-4, 4, -3, 3, 0] : 0,
      }}
      transition={isActive ? { scale: { ...SPRING_SNAPPY } } : { duration: 0.3 }}
    >
      <div className="flex items-center gap-3">
        <m.span
          className={`flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-bold ${iconBg} ${
            isActive ? 'll-active' : ''
          }`}
          animate={TERMINAL_STATUSES.has(status) ? { scale: [1.3, 1] } : {}}
          transition={SPRING_SNAPPY}
        >
          {iconChar}
        </m.span>
        <div className="min-w-0 flex-1">
          <div className="text-xs uppercase tracking-wider text-zinc-500">{title}</div>
          <div className={`truncate text-sm ${labelClass}`}>{label}</div>
        </div>
      </div>

      {/* Per-step inline progress bar */}
      {(isActive || status === 'passed' || status === 'settled' || status === 'blocked' || status === 'failed') && (
        <div className="mt-2 h-1 w-full overflow-hidden rounded-full bg-zinc-900">
          <div
            className={
              status === 'blocked' || status === 'failed'
                ? 'h-full bg-red-500 transition-all duration-300'
                : status === 'passed' || status === 'settled'
                  ? 'h-full bg-emerald-500 transition-all duration-300'
                  : 'll-bar-active h-full transition-all duration-200'
            }
            style={{ width: `${(activeFrac * 100).toFixed(1)}%` }}
          />
        </div>
      )}

      {/* Shimmer overlay only on active step */}
      {isActive && <div className="pointer-events-none absolute inset-0 ll-shimmer" />}
    </m.div>
  );
}
