# LedgerLens — Web App Dev Guide

Frontend for [LedgerLens](https://ledgerlens.gemsquared.ai/). Next.js 15 + React 19 + Tailwind 3.4 + TypeScript. Talks to a Go backend that ships as a **single binary** with this Next.js export embedded via `//go:embed`.

**You don't need to run the Go backend yourself.** It's already deployed at `https://ledgerlens.gemsquared.ai/`. The dev workflow points local Next.js dev at the hosted backend, so you can iterate on the UI without setting up Bright Data / GEM² / Anthropic API keys.

---

## Quickstart

```bash
git clone https://github.com/gem-squared/ledgerlens.git
cd ledgerlens/apps/web
pnpm install

# Point local dev at the hosted backend
cat > .env.local <<'EOF'
LEDGERLENS_BACKEND_URL=https://ledgerlens.gemsquared.ai
EOF

pnpm dev
# → opens on http://localhost:3001
```

Dev port is **3001** (not 3000) — see `package.json` `dev` script. Port 3000 was already in use elsewhere on the original dev machine.

The `LEDGERLENS_BACKEND_URL` env var feeds `next.config.mjs` rewrites: any request to `/api/v1/*` on `localhost:3001` is proxied to `https://ledgerlens.gemsquared.ai/api/v1/*` server-side, so the browser sees same-origin (no CORS hassle).

---

## API endpoints

Base URL: `https://ledgerlens.gemsquared.ai/api/v1/`

All endpoints are **public — no auth, no API key**. All responses are JSON (Content-Type `application/json`) unless noted.

### Reads

| Method | Path | Returns |
|---|---|---|
| `GET` | `/health` | `{status, service, settlement_mode, real_transaction_capability, cases}` |
| `GET` | `/cases` | `{cases: CaseListItem[]}` — Case A (blocked), Case B (approved) |
| `GET` | `/stats` | Aggregate dashboard stats — `dealsAudited`, `approved`, `blocked`, `avgAuditScore`, `simulatedSpendPreventedUSDC`, `auditGateUrl`, `auditGateAvgLatencyMs`, `modesBreakdown` |
| `GET` | `/audit-bundles` | `{bundles: BundleSummary[]}` — recent audit bundles |
| `GET` | `/audit-bundles/{decisionId}` | Full `DealRunResult` for one bundle (finalReport synthesized on read, mode/durationMs derived from timestamps) |

### Writes

| Method | Path | Body | Returns | Latency |
|---|---|---|---|---|
| `POST` | `/deals/run` | `{query, maxSpendUSDC?, requireGrounded?, mode?}` | `DealRunResult` (blocking) | ~20–45s |
| `POST` | `/deals/run-stream` | same | **SSE stream** of `StepEvent` then final `step: "result"` with full `DealRunResult` | ~20–45s |
| `POST` | `/cases/{a\|b}/run` | (no body) | `DealRunResult` (deterministic) | ~15–20s |

### Canonical response shape: `DealRunResult`

All three post-run paths (`/deals/run`, `/cases/*/run`, `/audit-bundles/{id}`) return the same shape — that's what makes `<RichRunResult>` reusable.

```ts
interface DealRunResult {
  mode: 'live' | 'replay' | 'prewarmed';
  judgeRequest: string;                    // the natural-language request
  buyerIntent: BuyerIntent;
  agentNarrative: string[];
  evidenceReceipts: EvidenceReceipt[];
  sellerOffer: SellerOffer;
  decision: DecisionPacket;                // verdict, claimAssessments, reason
  settlement: SimulatedSettlement;
  l1?: GateResponse;                       // L1 P-check (verdict, score, reasons[])
  l2?: GateResponse;                       // L2 O-check (verdict, score, reasons[])
  bundlePath: string;
  finalReport: FinalReport;                // human-readable summary
  durationMs: number;
}
```

Full types in [`apps/web/lib/types.ts`](./lib/types.ts).

---

## Quick curl smoke

```bash
# Health
curl -sS https://ledgerlens.gemsquared.ai/api/v1/health | jq

# List recent audit bundles
curl -sS https://ledgerlens.gemsquared.ai/api/v1/audit-bundles | jq '.bundles[0]'

# Pull one bundle (use a decisionId from the list above)
curl -sS https://ledgerlens.gemsquared.ai/api/v1/audit-bundles/dp_<id> | jq '.finalReport'

# Run Case B (approved, ~20s)
curl -sS -X POST https://ledgerlens.gemsquared.ai/api/v1/cases/b/run -m 60 | jq '.finalReport.headline'

# Stats
curl -sS https://ledgerlens.gemsquared.ai/api/v1/stats | jq

# LIVE deal end-to-end (warning: ~30–45s blocking)
curl -sS -X POST https://ledgerlens.gemsquared.ai/api/v1/deals/run \
  -H 'Content-Type: application/json' \
  -m 90 \
  -d '{"query":"Find a trustworthy live NYSE + NASDAQ market data provider under $0.001/query.","maxSpendUSDC":0.001,"requireGrounded":true}' \
  | jq '.finalReport.headline'
```

---

## Caveats

### LIVE deals through `pnpm dev`'s proxy time out at ~30s

A LIVE deal typically takes **30–45s** end-to-end (Bright Data fetch + GEM² L1+L2 audit). The Next.js dev-mode rewrite proxy closes the upstream socket at ~30s, so the streamed result never reaches your browser. **This is a Next.js dev-mode limitation, not a LedgerLens bug.**

Workarounds while testing locally:
- **Use REPLAY mode** (Case A/B) — completes in ~15–20s, never times out.
- **Use the View modal** (click any row in Recent Audit Samples) — those are blocking GETs that finish in ~50ms.
- **Test LIVE on production:** `https://ledgerlens.gemsquared.ai/` runs against the same backend with no Next.js proxy → no timeout.

If you really need to drive the LIVE path locally, use curl/fetch directly against the hosted backend from your terminal (the dev server isn't in the path).

### No backend env vars needed

The hosted backend has all the Bright Data / GEM² / Anthropic keys. You're just consuming the API.

### Production deploys are atomic and owned upstream

`make deploy` (run from project root by David) chains: `pnpm build --output=export` → embed into Go binary → cross-compile linux/amd64 → SCP to VPS → backup-ladder rotate → systemctl restart → health check → auto-rollback on failure. So when you ship a UI change, **David** rebuilds + redeploys. There's no separate web-app deploy — the web app **is** part of the Go binary.

---

## What you can iterate on

- UI components: `apps/web/app/components/*.tsx`
- Page composition: `apps/web/app/page.tsx`
- Tailwind styles + custom keyframes: `apps/web/app/globals.css`
- Tailwind config: `apps/web/tailwind.config.ts` (note the `simBadge` color = LedgerLens accent indigo)
- Client-side types: `apps/web/lib/types.ts` — keep aligned with `packages/contracts-ts/types.ts` (auto-generated from Go via `make schemas`)
- SSE consumer: `apps/web/lib/sse.ts`
- API client: `apps/web/lib/api.ts`

### Don't touch (owned upstream / regenerated)

- `cmd/ledgerlens/web_static/` — generated by `make web-export`, gitignored except a placeholder `index.html`
- `packages/contracts-ts/types.ts` — regenerated from Go structs via `make schemas`
- Backend Go code in `cmd/` and `internal/` — if you need a new endpoint or response field, open an issue first

---

## How to ship a UI change

```bash
# 1. Branch from main
git checkout main && git pull
git checkout -b feature/your-change

# 2. Edit. Test locally against the hosted backend.
pnpm dev   # http://localhost:3001

# 3. Commit + push
git push origin feature/your-change

# 4. Open a PR against main. David reviews, merges,
#    and runs `make deploy` to roll prod forward.
```

That's it. Welcome aboard.

— gem-squared · david@gineers.ai
