'use client';

import { useEffect, useState } from 'react';
import { getAuditBundle } from '@/lib/api';
import type { DealRunResult } from '@/lib/types';
import { RichRunResult } from './RichRunResult';

interface BundleViewerProps {
  decisionId: string;
  onClose: () => void;
}

// Modal-style overlay that fetches a persisted audit bundle and renders it
// using the same RichRunResult surface a fresh LIVE deal produces. The raw
// JSON is preserved under a collapsible details block at the bottom of the
// modal for judges who want forensic proof.
export function BundleViewer({ decisionId, onClose }: BundleViewerProps) {
  const [bundle, setBundle] = useState<DealRunResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const data = await getAuditBundle(decisionId);
        if (!cancelled) setBundle(data);
      } catch (err) {
        if (!cancelled) setError(err instanceof Error ? err.message : String(err));
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [decisionId]);

  // ESC closes
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') onClose();
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/70 p-4"
      onClick={onClose}
    >
      <div
        className="relative max-h-[90vh] w-full max-w-5xl overflow-hidden rounded-xl border border-zinc-700 bg-zinc-950 shadow-2xl"
        onClick={(e) => e.stopPropagation()}
      >
        <header className="flex items-center justify-between gap-3 border-b border-zinc-800 px-5 py-3">
          <div>
            <div className="text-[10px] font-semibold uppercase tracking-wider text-zinc-500">
              Audit Bundle
            </div>
            <div className="font-mono text-sm text-zinc-200">{decisionId}</div>
          </div>
          <div className="flex items-center gap-2">
            <a
              href={`/api/v1/audit-bundles/${decisionId}`}
              target="_blank"
              rel="noopener noreferrer"
              className="rounded border border-zinc-700 px-2 py-1 text-[11px] text-zinc-300 hover:border-zinc-500"
            >
              raw JSON ↗
            </a>
            <button
              type="button"
              onClick={onClose}
              className="rounded border border-zinc-700 px-2 py-1 text-[11px] text-zinc-300 hover:border-zinc-500"
            >
              close (esc)
            </button>
          </div>
        </header>
        <div className="max-h-[78vh] overflow-auto p-5">
          {error && (
            <div className="rounded border border-red-700/50 bg-red-900/20 p-3 text-sm text-red-300">
              <strong>Error:</strong> {error}
            </div>
          )}
          {!error && !bundle && (
            <p className="text-xs text-zinc-500">Loading bundle…</p>
          )}
          {bundle && (
            <>
              <RichRunResult result={bundle} />

              <details className="mt-6 rounded-lg border border-zinc-800 bg-zinc-900/40">
                <summary className="cursor-pointer select-none px-4 py-2 text-[11px] font-semibold uppercase tracking-wider text-zinc-400 hover:bg-zinc-900/60">
                  Show raw audit-bundle JSON (forensic)
                </summary>
                <pre className="overflow-x-auto whitespace-pre-wrap break-words border-t border-zinc-800 p-4 font-mono text-[11px] leading-relaxed text-zinc-400">
                  {JSON.stringify(bundle, null, 2)}
                </pre>
              </details>
            </>
          )}
        </div>
      </div>
    </div>
  );
}
