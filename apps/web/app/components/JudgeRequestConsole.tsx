'use client';

import { useEffect, useRef, useState } from 'react';
import { runDealStream } from '@/lib/sse';
import { listCases, runCase } from '@/lib/api';
import type { CaseListItem, DealRunResult, RunMode, StepEvent } from '@/lib/types';
import { AgentFlowTimeline } from './AgentFlowTimeline';
import { RichRunResult } from './RichRunResult';

const DEFAULT_QUERY =
  'Find a trustworthy live NYSE + NASDAQ market data provider under $0.001/query.';

// Curated LIVE queries that exercise distinct LedgerLens behaviors. Used
// only to populate the textarea — the judge can edit before running, and
// the actual deal still runs LIVE through Bright Data + GEM².
const SAMPLE_QUERIES: { label: string; query: string }[] = [
  {
    label: 'APPROVED CASE — NYSE + NASDAQ live market data',
    query: 'Find a trustworthy live NYSE + NASDAQ market data provider under $0.001/query.',
  },
  {
    label: 'APPROVED CASE — BTC-USD live spot price feed',
    query: 'Find a real-time BTC-USD spot price feed under $0.0005/query.',
  },
  {
    label: 'BLOCKED CASE — Live political polling aggregator',
    query: 'Find a live political polling aggregator updated within 1 hour, free tier.',
  },
  {
    label: 'BLOCKED CASE — Real-time satellite imagery API',
    query: 'Find a real-time satellite imagery API with 10-meter resolution under $0.01/query.',
  },
  {
    label: 'OFF-DOMAIN CASE — Dinner recipes',
    query: 'Find me dinner recipes for tonight.',
  },
];

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
  const [cases, setCases] = useState<CaseListItem[]>([]);

  // The sample-picker dropdown animates a typing placeholder until the
  // judge interacts with it (focus / mousedown / change). After that,
  // animation stops permanently.
  const [sampleInteracted, setSampleInteracted] = useState(false);
  const sampleSelectRef = useRef<HTMLSelectElement | null>(null);

  const abortRef = useRef<AbortController | null>(null);

  // Force a re-render every 250ms while running, so the elapsed-time
  // display + heartbeat detection stay live even with no incoming events.
  const [, forceTick] = useState(0);
  useEffect(() => {
    if (!running) return;
    const id = setInterval(() => forceTick((n) => (n + 1) % 1_000_000), 250);
    return () => clearInterval(id);
  }, [running]);

  // Load the canonical Case A / Case B list once — used by REPLAY mode.
  useEffect(() => {
    listCases().then(setCases).catch((err) => {
      console.error('listCases failed', err);
    });
  }, []);

  // Typing-effect placeholder on the sample-picker dropdown. Mutates
  // the first <option>'s visible text char-by-char with a blinking
  // caret until the judge interacts. Native <select> placeholder text
  // can't be animated any other way; this is the canonical trick.
  useEffect(() => {
    if (sampleInteracted) return;
    if (mode === 'replay') return; // dropdown only renders in LIVE/PRE-WARMED
    const select = sampleSelectRef.current;
    if (!select || !select.options[0]) return;

    const fullText = '▾ Pick a sample request — judge can edit before running';
    let charIndex = 0;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout> | null = null;

    const tick = () => {
      if (cancelled) return;
      charIndex++;
      if (charIndex > fullText.length) {
        // Pause at full text, then restart from the beginning.
        timer = setTimeout(() => {
          if (cancelled) return;
          charIndex = 0;
          tick();
        }, 1800);
        return;
      }
      if (select.options[0]) {
        // Alternate caret visibility for a blink without an extra timer.
        const caret = charIndex % 2 === 0 ? '▌' : ' ';
        select.options[0].text = fullText.substring(0, charIndex) + caret;
      }
      timer = setTimeout(tick, 55);
    };
    tick();

    return () => {
      cancelled = true;
      if (timer) clearTimeout(timer);
      // Restore the full descriptive text so the open dropdown reads cleanly.
      if (select.options[0]) {
        select.options[0].text =
          '▾ Pick a sample request (populates the textarea — judge can edit before running)';
      }
    };
  }, [sampleInteracted, mode]);

  const stopSampleAnimation = () => {
    if (!sampleInteracted) setSampleInteracted(true);
  };

  function reset() {
    setEvents([]);
    setLastEventAt(null);
    setResult(null);
    setError(null);
    setPoliteReject(null);
  }

  function onRunLive() {
    if (running) return;
    reset();
    setRunning(true);
    setRunStartedAt(Date.now());
    abortRef.current = runDealStream(
      {
        query,
        maxSpendUSDC: 0.001,
        requireGrounded: true,
        mode: 'live',
      },
      {
        onStep: (e) => {
          setEvents((prev) => [...prev, e]);
          setLastEventAt(Date.now());
          if (e.status === 'rejected') {
            const d = e.detail as { polite_reject?: string } | undefined;
            if (d?.polite_reject) setPoliteReject(d.polite_reject);
          }
        },
        onResult: (r) => {
          setResult(r);
          setRunning(false);
          onRunComplete?.();
        },
        onError: (err) => {
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
          setRunning(false);
        },
      },
    );
  }

  async function onRunReplay(caseId: string) {
    if (running) return;
    reset();
    setRunning(true);
    setRunStartedAt(Date.now());
    try {
      const r = await runCase(caseId);
      setResult(r);
      onRunComplete?.();
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : String(e));
    } finally {
      setRunning(false);
    }
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

        {/* LIVE / PRE-WARMED: free-text query → SSE deal pipeline */}
        {mode !== 'replay' && (
          <>
            <div className="mb-2 flex items-center gap-2">
              <label
                htmlFor="sample-picker"
                className="flex items-center gap-1 whitespace-nowrap text-[11px] font-semibold uppercase tracking-wider"
              >
                <span
                  className={sampleInteracted ? 'inline-block' : 'll-attention-bob'}
                  aria-hidden
                >✨</span>
                <span
                  className={sampleInteracted ? 'text-zinc-500' : 'll-attention-shimmer'}
                >Sample requests</span>
              </label>
              <select
                ref={sampleSelectRef}
                id="sample-picker"
                disabled={running}
                value=""
                onFocus={stopSampleAnimation}
                onMouseDown={stopSampleAnimation}
                onChange={(e) => {
                  stopSampleAnimation();
                  const pick = SAMPLE_QUERIES.find((s) => s.label === e.target.value);
                  if (pick) setQuery(pick.query);
                  // Reset back to placeholder so re-picking the same item still fires onChange.
                  e.target.value = '';
                }}
                className="flex-1 rounded-md border border-zinc-800 bg-zinc-950 px-2 py-1 text-xs text-zinc-300 focus:border-simBadge focus:outline-none disabled:opacity-50"
              >
                <option value="">▾ Pick a sample request (populates the textarea — judge can edit before running)</option>
                {SAMPLE_QUERIES.map((s) => (
                  <option key={s.label} value={s.label}>{s.label}</option>
                ))}
              </select>
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
            </div>

            <div className="mt-4 flex flex-wrap items-center gap-3">
              <button
                type="button"
                onClick={onRunLive}
                disabled={running || !query.trim() || mode === 'prewarmed'}
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
          </>
        )}

        {/* REPLAY: canonical Case A / Case B cards, click-to-run inline */}
        {mode === 'replay' && (
          <>
            <div className="rounded-md border border-zinc-800 bg-zinc-950 px-3 py-2 text-xs leading-relaxed text-zinc-400">
              Replay mode runs a fixed scripted scenario through the same GEM² trust gate.
              No live web variance — useful as a demo safety net if the live audit gate is
              slow or unreachable. Click a card to run.
            </div>

            <div className="mt-3 grid gap-3 sm:grid-cols-2">
              {cases.map((c) => (
                <button
                  key={c.id}
                  type="button"
                  onClick={() => onRunReplay(c.id)}
                  disabled={running}
                  className={`group flex flex-col gap-2 rounded-lg border p-4 text-left transition ${
                    running
                      ? 'cursor-not-allowed border-zinc-800 bg-zinc-900/40 opacity-60'
                      : 'border-zinc-800 bg-zinc-900/40 hover:border-simBadge hover:bg-zinc-900'
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <span className="font-mono text-xs uppercase text-zinc-500">▶ run</span>
                    <h3 className="text-base font-semibold text-zinc-100">{c.title}</h3>
                  </div>
                  <p className="text-xs leading-relaxed text-zinc-400">{c.description}</p>
                </button>
              ))}
            </div>

            {!running && result && (
              <div className="mt-4">
                <button
                  type="button"
                  onClick={onReset}
                  className="rounded-md border border-zinc-700 px-3 py-2 text-xs text-zinc-300 hover:border-zinc-500"
                >
                  Reset
                </button>
              </div>
            )}
          </>
        )}

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

      {/* ── Live timeline (LIVE/PRE-WARMED only; REPLAY is blocking) ───── */}
      {mode !== 'replay' && runStartedAt !== null && (
        <section className="rounded-xl border border-zinc-800 bg-zinc-900/30 p-5">
          <AgentFlowTimeline
            mode={mode}
            runStartedAt={runStartedAt}
            events={events}
            done={!running && (result !== null || error !== null || politeReject !== null)}
            lastEventAt={lastEventAt}
          />
        </section>
      )}

      {/* Unified post-run surface — same UI for LIVE, REPLAY, and Recent-Activity View. */}
      {result && <RichRunResult result={result} />}
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
