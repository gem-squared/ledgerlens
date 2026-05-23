// SSE consumer for POST /api/v1/deals/run-stream.
//
// EventSource is GET-only, so we use fetch + ReadableStream. The server
// emits frames in the standard format:
//   event: <step|result|error>
//   data: <single-line JSON>
//   <blank line>
//
// Each frame is parsed and dispatched to one of three callbacks.

import type { StepEvent, DealRunResult } from './types';

export interface SSECallbacks {
  onStep: (e: StepEvent) => void;
  onResult: (r: DealRunResult) => void;
  onError: (err: { error: string }) => void;
  onClose?: () => void;
}

export interface RunDealStreamBody {
  query: string;
  maxSpendUSDC?: number;
  requireGrounded?: boolean;
  mode?: 'live' | 'prewarmed';
}

/** runDealStream POSTs to the SSE endpoint and pumps events into callbacks.
 *  Returns an AbortController so callers can cancel the run mid-stream. */
export function runDealStream(body: RunDealStreamBody, cb: SSECallbacks): AbortController {
  const ac = new AbortController();
  (async () => {
    try {
      const r = await fetch('/api/v1/deals/run-stream', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
        signal: ac.signal,
      });
      if (!r.ok) {
        const text = await r.text();
        cb.onError({ error: `HTTP ${r.status}: ${text}` });
        return;
      }
      if (!r.body) {
        cb.onError({ error: 'no response body' });
        return;
      }
      const reader = r.body.getReader();
      const decoder = new TextDecoder();
      let buffer = '';
      while (true) {
        const { done, value } = await reader.read();
        if (done) break;
        buffer += decoder.decode(value, { stream: true });
        let sep;
        // SSE frames separated by \n\n
        while ((sep = buffer.indexOf('\n\n')) >= 0) {
          const frame = buffer.slice(0, sep);
          buffer = buffer.slice(sep + 2);
          parseFrame(frame, cb);
        }
      }
      // flush any tail frame
      if (buffer.trim()) parseFrame(buffer, cb);
      cb.onClose?.();
    } catch (err: unknown) {
      const msg = err instanceof Error ? err.message : String(err);
      if (msg.includes('aborted') || msg.includes('AbortError')) {
        cb.onClose?.();
        return;
      }
      cb.onError({ error: msg });
    }
  })();
  return ac;
}

function parseFrame(frame: string, cb: SSECallbacks): void {
  if (frame.startsWith(':')) return; // SSE comment / keep-alive
  let evType = 'message';
  let dataStr = '';
  for (const line of frame.split('\n')) {
    if (line.startsWith('event: ')) evType = line.slice(7).trim();
    else if (line.startsWith('event:')) evType = line.slice(6).trim();
    else if (line.startsWith('data: ')) dataStr += line.slice(6);
    else if (line.startsWith('data:')) dataStr += line.slice(5);
  }
  if (!dataStr) return;
  let data: unknown;
  try {
    data = JSON.parse(dataStr);
  } catch {
    return;
  }
  if (evType === 'step') cb.onStep(data as StepEvent);
  else if (evType === 'result') cb.onResult(data as DealRunResult);
  else if (evType === 'error') cb.onError(data as { error: string });
}
