import type { EvidenceReceipt } from '@/lib/types';

const PRODUCT_COLOR: Record<string, string> = {
  SERP:     'bg-blue-500/20 text-blue-300',
  UNLOCKER: 'bg-purple-500/20 text-purple-300',
  BROWSER:  'bg-green-500/20 text-green-300',
  SCRAPER:  'bg-orange-500/20 text-orange-300',
  MCP:      'bg-pink-500/20 text-pink-300',
};

export function EvidenceList({ receipts }: { receipts: EvidenceReceipt[] }) {
  if (!receipts?.length) return <p className="text-sm text-zinc-500">No evidence receipts.</p>;
  return (
    <ul className="space-y-2">
      {receipts.map((r) => (
        <li key={r.receiptId} className="rounded border border-zinc-800 bg-zinc-900/40 p-3 text-sm">
          <div className="flex items-center gap-2">
            <span
              className={`rounded px-1.5 py-0.5 font-mono text-[10px] font-bold ${
                PRODUCT_COLOR[r.brightDataProduct] ?? 'bg-zinc-700 text-zinc-200'
              }`}
            >
              {r.brightDataProduct}
            </span>
            <code className="font-mono text-xs text-zinc-500">{r.receiptId}</code>
            <span className="text-xs text-zinc-500">{r.fetchedAt}</span>
          </div>
          <div className="mt-1 truncate text-xs text-zinc-400" title={r.url}>
            {r.url}
          </div>
          <div className="mt-1 truncate font-mono text-[10px] text-zinc-600" title={r.contentHash}>
            {r.contentHash}
          </div>
        </li>
      ))}
    </ul>
  );
}
