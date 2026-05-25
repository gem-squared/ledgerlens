'use client';

import { useEffect, useState } from 'react';
import { listCases, runCase } from '@/lib/api';
import type { CaseListItem, DealRunResult } from '@/lib/types';
import { RichRunResult } from './RichRunResult';

export function CaseRunner() {
  const [cases, setCases] = useState<CaseListItem[]>([]);
  const [running, setRunning] = useState<string | null>(null);
  const [result, setResult] = useState<DealRunResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    listCases().then(setCases).catch((e) => setError(String(e)));
  }, []);

  async function onRun(id: string) {
    setRunning(id);
    setError(null);
    setResult(null);
    try {
      const r = await runCase(id);
      setResult(r);
    } catch (e: unknown) {
      setError(String(e));
    } finally {
      setRunning(null);
    }
  }

  return (
    <div className="space-y-8">
      <section className="grid gap-4 sm:grid-cols-2">
        {cases.map((c) => (
          <button
            key={c.id}
            onClick={() => onRun(c.id)}
            disabled={running !== null}
            className={`group flex flex-col gap-2 rounded-lg border p-5 text-left transition ${
              running === c.id
                ? 'border-simBadge bg-simBadge/10'
                : 'border-zinc-800 bg-zinc-900/40 hover:border-zinc-600 hover:bg-zinc-900'
            } disabled:cursor-not-allowed disabled:opacity-60`}
          >
            <div className="flex items-center gap-2">
              <span className="font-mono text-xs uppercase text-zinc-500">▶ run</span>
              <h2 className="text-lg font-semibold">{c.title}</h2>
              {running === c.id && (
                <span className="ml-auto text-xs text-simBadge">running… (~15s)</span>
              )}
            </div>
            <p className="text-sm text-zinc-400">{c.description}</p>
          </button>
        ))}
      </section>

      {error && (
        <div className="rounded-lg border border-red-700/50 bg-red-900/20 p-4 text-sm text-red-300">
          <strong>Error:</strong> {error}
        </div>
      )}

      {result && <RichRunResult result={result} />}
    </div>
  );
}
