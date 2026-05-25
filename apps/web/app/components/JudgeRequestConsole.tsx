'use client';

import { useEffect, useRef, useState } from 'react';
import { runDealStream } from '@/lib/sse';
import type { DealRunResult, RunMode, StepEvent } from '@/lib/types';
import { AgentFlowTimeline } from './AgentFlowTimeline';
import { FinalReportPanel } from './FinalReportPanel';
import { AuditScoreCard } from './AuditScoreCard';
import { ClaimAssessmentTable } from './ClaimAssessmentTable';
import { SettlementCard } from './SettlementCard';
import { EvidenceList } from './EvidenceList';
import { VerdictReveal } from './VerdictReveal';

const DEFAULT_QUERY =
  'Find a trustworthy live NYSE + NASDAQ market data provider under $0.001/query.';

interface JudgeRequestConsoleProps {
  /** Called when a run completes (success or error) so the dashboard can refetch. */
  onRunComplete?: () => void;
}

export function JudgeRequestConsole({ onRunComplete }: JudgeRequestConsoleProps = {}) {
  const [query, setQuery] = useState(DEFAULT_QUERY);
  const [mode, setMode] = useState<RunMode>('live');
  const [running, setRunning] = useState(false);
  const [runStartedAt, setRunStartedAt] = useState<number | null>(null);
  const [events, setEvents] = useState<StepEvent[]>([]);
  const [lastEventAt, setLastEventAt] = useState<number | null>(null);
  const [result, setResult] = useState<DealRunResult | null>(null);
  const [error, setError] = useState<string | null>(null);
  const [politeReject, setPoliteReject] = useState<string | null>(null);

  const abortRef = useRef<AbortController | null>(null);

  // Force a re-render every 250ms while running, so the elapsed-time
  // display + heartbeat detection stay live even with no incoming events.
  const [, forceTick] = useState(0);
  useEffect(() => {
    if (!running) return;
    const id = setInterval(() => forceTick((n) => (n + 1) % 1_000_000), 250);
    return () => clearInterval(id);
  }, [running]);

  function reset() {
    setEvents([]);
    setLastEventAt(null);
    setResult(null);
    setError(null);
    setPoliteReject(null);
  }

  function onRun() {
    if (running) return;
    reset();
    setRunning(true);
    setRunStartedAt(Date.now());
    abortRef.current = runDealStream(
      {
        query,
        maxSpendUSDC: 0.001,
        requireGrounded: true,
        mode: mode === 'replay' ? 'live' : mode,
      },
      {
        onStep: (e) => {
          setEvents((prev) => [...prev, e]);
          setLastEventAt(Date.now());
          // Off-domain reject surfaces as a step event with status=rejected.
          if (e.status === 'rejected') {
            const d = e.detail as { polite_reject?: string } | undefined;
            if (d?.polite_reject) setPoliteReject(d.polite_reject);
          }
        },
        onResult: (r) => {
          setResult(r);
          setRunning(false);
          // Tell the dashboard a new audit bundle exists → refetch counters.
          onRunComplete?.();
        },
        onError: (err) => {
          // Parse off_domain rejection from the JSON path (race condition guard)
          if (err.error?.includes('off_domain')) {
            setPoliteReject(
              'This demo is scoped to autonomous web-data purchase requests. Try the default NYSE/NASDAQ market data request.',
            );
          } else {
            setError(err.error || 'unknown error');
          }
          setRunning(false);
        },
        onClose: () => {
          // ensure we drop the running flag even if onResult wasn't reached
          setRunning(false);
        },
      },
    );
  }

  function onCancel() {
    abortRef.current?.abort();
    setRunning(false);
  }

  function onReset() {
    reset();
    setRunStartedAt(null);
    setQuery(DEFAULT_QUERY);
  }

  return (
    <div className="space-y-6">
      {/* ── Console ────────────────────────────────────────────────────── */}
      <section className="rounded-xl border border-zinc-800 bg-zinc-900/50 p-5">
        <div className="mb-3 flex items-baseline justify-between gap-4">
          <h2 className="text-base font-semibold uppercase tracking-wider text-zinc-300">
            Judge Request Console
          </h2>
          <ModeSwitch mode={mode} setMode={setMode} disabled={running} />
        </div>

        <textarea
          className="block h-24 w-full resize-none rounded-lg border border-zinc-800 bg-zinc-950 px-3 py-2 text-sm leading-relaxed text-zinc-100 placeholder-zinc-600 focus:border-simBadge focus:outline-none disabled:opacity-50"
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          disabled={running}
          spellCheck={false}
        />

        <div className="mt-3 rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-xs leading-relaxed text-zinc-400">
          {mode === 'live' && (
            <>
              Live verification typically completes in <strong>20–45 seconds</strong>. Each step
              below updates as Bright Data fetches evidence and GEM² audits the seller&apos;s claim.{' '}
              <span className="italic text-zinc-300">
                Fast agents are dangerous if they spend before verification. LedgerLens deliberately waits.
              </span>
            </>
          )}
          {mode === 'prewarmed' && (
            <>
              Pre-warmed mode reuses cached Bright Data evidence and runs a live GEM² audit. Target
              latency <strong>5–8 seconds</strong>. The trust gate is still real.{' '}
              <span className="rounded bg-yellow-500/20 px-1.5 py-0.5 text-[10px] font-semibold uppercase tracking-wider text-yellow-300">
                Slice 3 — not yet shipped
              </span>
            </>
          )}
          {mode === 'replay' && (
            <>
              Replay mode shows a fixed deterministic Case A (blocked) or Case B (approved) — no
              live verification. Use the cards at the bottom of the page.
            </>
          )}
        </div>

        <div className="mt-4 flex flex-wrap items-center gap-3">
          <button
            type="button"
            onClick={onRun}
            disabled={running || mode === 'replay' || !query.trim()}
            className="rounded-md bg-simBadge px-5 py-2.5 text-sm font-semibold text-white shadow-md hover:bg-indigo-500 disabled:cursor-not-allowed disabled:opacity-40"
          >
            {running ? 'Verifying…' : '▸ Run Autonomous Deal'}
          </button>
          {running && (
            <button
              type="button"
              onClick={onCancel}
              className="rounded-md border border-zinc-700 px-3 py-2 text-xs text-zinc-300 hover:border-zinc-500"
            >
              Cancel
            </button>
          )}
          {!running && (events.length > 0 || result) && (
            <button
              type="button"
              onClick={onReset}
              className="rounded-md border border-zinc-700 px-3 py-2 text-xs text-zinc-300 hover:border-zinc-500"
            >
              Reset
            </button>
          )}
          <span className="text-[11px] uppercase tracking-wider text-zinc-500">
            Mode: <span className="text-zinc-300">{mode}</span>
          </span>
        </div>

        {politeReject && (
          <div className="mt-4 rounded-lg border border-amber-500/40 bg-amber-500/10 p-3 text-sm text-amber-200">
            <strong className="block uppercase tracking-wider text-amber-300">
              Off-domain query
            </strong>
            <p className="mt-1">{politeReject}</p>
          </div>
        )}

        {error && (
          <div className="mt-4 rounded-lg border border-red-700/50 bg-red-900/20 p-3 text-sm text-red-300">
            <strong>Error:</strong> {error}
          </div>
        )}
      </section>

      {/* ── Live timeline ──────────────────────────────────────────────── */}
      {runStartedAt !== null && (
        <section className="rounded-xl border border-zinc-800 bg-zinc-900/30 p-5">
          <AgentFlowTimeline
            mode={mode === 'replay' ? 'live' : mode}
            runStartedAt={runStartedAt}
            events={events}
            done={!running && (result !== null || error !== null || politeReject !== null)}
            lastEventAt={lastEventAt}
          />
        </section>
      )}

      {/* ── Layer 0: Final Report banner (always visible) ──────────────── */}
      {/* ── Layer 1: Audit Score rings + Layer 2/3 drilldowns ──────────── */}
      {/* ── Settlement + Evidence summaries ────────────────────────────── */}
      {result && (
        <>
          <VerdictReveal
            verdict={
              result.decision?.verdict === 'APPROVED_BY_TRUST_GATE'
                ? 'approved'
                : 'blocked'
            }
          />
          <FinalReportPanel report={result.finalReport} durationMs={result.durationMs} />

          <AuditScoreCard l1={result.l1} l2={result.l2} />

          <div className="grid gap-6 lg:grid-cols-2">
            <SettlementCard settlement={result.settlement} />
            <section className="rounded-xl border border-zinc-800 bg-zinc-900/30 p-5">
              <h3 className="mb-3 flex items-baseline justify-between text-xs font-semibold uppercase tracking-wider text-zinc-300">
                <span>Bright Data Evidence</span>
                <span className="font-mono text-zinc-500">
                  {result.evidenceReceipts?.length ?? 0} receipts
                </span>
              </h3>
              <EvidenceList receipts={result.evidenceReceipts} />
            </section>
          </div>

          {/* Technical details (claim assessments) — collapsed by default */}
          <details className="group rounded-xl border border-zinc-800 bg-zinc-900/30 open:bg-zinc-900/40">
            <summary className="cursor-pointer select-none list-none px-5 py-3 text-sm text-zinc-300 hover:bg-zinc-900/60">
              <span className="inline-block w-4 text-zinc-500 transition-transform group-open:rotate-90">›</span>
              Show claim-by-claim assessment ({result.decision?.claimAssessments?.length ?? 0} claims · canonical EEF tags)
            </summary>
            <div className="border-t border-zinc-800 p-5">
              <ClaimAssessmentTable claims={result.decision?.claimAssessments ?? []} />
            </div>
          </details>

          <div className="flex flex-wrap gap-x-6 gap-y-1 text-xs text-zinc-500">
            <span>mode: <code>{result.mode}</code></span>
            <span>duration: {result.durationMs} ms</span>
            <span>bundle: <code className="break-all">{result.bundlePath}</code></span>
          </div>
        </>
      )}
    </div>
  );
}

// ── Three-mode toggle ────────────────────────────────────────────────────

function ModeSwitch({
  mode,
  setMode,
  disabled,
}: {
  mode: RunMode;
  setMode: (m: RunMode) => void;
  disabled: boolean;
}) {
  const items: { id: RunMode; label: string; sub: string }[] = [
    { id: 'live',      label: 'LIVE',       sub: '20–45s' },
    { id: 'prewarmed', label: 'PRE-WARMED', sub: '5–8s' },
    { id: 'replay',    label: 'REPLAY',     sub: 'instant' },
  ];
  return (
    <div className="flex gap-1.5">
      {items.map((it) => {
        const active = mode === it.id;
        return (
          <button
            key={it.id}
            type="button"
            onClick={() => !disabled && setMode(it.id)}
            disabled={disabled}
            className={`rounded-md border px-2.5 py-1 text-[11px] font-mono uppercase tracking-wider transition ${
              active
                ? 'border-simBadge bg-simBadge/20 text-white'
                : 'border-zinc-800 text-zinc-500 hover:border-zinc-600 hover:text-zinc-300'
            } disabled:cursor-not-allowed disabled:opacity-50`}
            title={`${it.label} · ${it.sub}`}
          >
            {it.label} <span className="opacity-60">· {it.sub}</span>
          </button>
        );
      })}
    </div>
  );
}
