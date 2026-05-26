# LedgerLens

> **No grounded claim, no payment.**

![LedgerLens — autonomous deal with trust-gated x402 settlement](./assets/demo.gif)

LedgerLens is an **x402-native Agent-to-Agent Payments product** with a **Bright Data-powered web evidence layer** and a **GEM² Trust Gate** before settlement.

Buyer agents need fresh web data. Seller agents claim they have it. LedgerLens uses Bright Data to verify those claims against live public web evidence, then asks GEM² to audit the claims through L1 P-check + L2 O-check before any payment is allowed. Only if the claim is sufficiently grounded does the system permit x402 settlement.

**Built for:** [Bright Data Web Data UNLOCKED Hackathon](https://lablab.ai/ai-hackathons/brightdata-ai-agents-web-data-hackathon) (2026-05-25 → 2026-05-31).
**Tracks:** Agent-to-Agent Payments (primary) · Web Data Commerce · General Web Data.
**Authors:** Inseok "David" Seo (`david@gineers.ai`) — GEM².AI.

---

## What it is

```
Buyer Agent requests live web data
   ↓
Seller Agent offers it with a claim
   ↓
LedgerLens discovers candidate offers
   ↓
Bright Data fetches public evidence (SERP + Unlocker + Browser + MCP)
   ↓
GEM² L1 P-check + L2 O-check (gem2-tpmn-checker.fly.dev)
   ↓
GEM² L3 Trust Gate composes the verdict
   ↓
If APPROVED  → simulated x402 settlement receipt
If BLOCKED   → no payment authorization issued
   ↓
Audit bundle exported (regulator-replay-ready)
```

The visible L2 audit panel — with `[EEF-⊢/⊨/⊬/⊥]` tags, `[SPT-…]` overclaim guards, and `[EVIDENCE-N]` evidence correlation — is the originality moat. No other team will ship "the agent must prove its claim is grounded before it spends money."

---

## Pitch sentence

> **"For demo safety, we do not connect a personal Coinbase account or expose private keys. Instead, we simulate the x402 payment lifecycle and focus on the core innovation: the GEM² Trust Gate decides whether an autonomous agent is allowed to pay."**

Shorter:

> **"We simulate settlement, but the trust gate is real."**

---

## x402 Demo Mode

LedgerLens demonstrates an x402-compatible payment flow in **simulation mode**.

For public demo safety, **no real private keys, Coinbase accounts, or mainnet/testnet funds are used**. The system models the x402 payment lifecycle:

1. Payment required (HTTP 402-shaped challenge)
2. Bright Data evidence collection
3. GEM² L1 P-check + L2 O-check via the deployed `gem2-tpmn-checker.fly.dev` audit-gate API
4. L3 Trust Gate decision (composite verdict)
5. Payment authorization or block
6. Simulated settlement receipt (tagged `mode: "simulation"`, `real_transaction: false`)

The purpose of the demo is to prove the **governance layer before settlement**:

> **No grounded claim, no payment.**

A real x402 settlement implementation (Coinbase Go SDK on Base Sepolia → mainnet) is a post-hackathon swap behind the `Settler` interface. The Trust Gate code does not change.

---

## Architecture

| Layer | Role | Implementation |
|---|---|---|
| **Evidence Layer** | Forensic acquisition of source material via Bright Data | `internal/brightdata/` — SERP + Unlocker + Browser + MCP wrappers; receipts in `artifacts/fetch_receipts/` |
| **Memory Layer** | Entity, offer, seller, buyer, source memory | `internal/trustgate/memory/` — SQLite entity graph |
| **Verification Layer** | L1 P-check + L2 O-check via production audit gate | `internal/trustgate/auditgate/` calling `gem2-tpmn-checker.fly.dev` |
| **Release Layer** | Composite policy gate before settlement | `internal/trustgate/release/` + `internal/paymentgate/` |
| **Settlement** | x402 protocol **simulation** (no chain) | `internal/paymentgate/simulated_settler.go` |

---

## Run locally

Requires Go ≥ 1.22, Node ≥ 20, pnpm ≥ 8.

```bash
# 1) Backend (Go binary)
cp .env.example .env
# fill in: GEM2_API_KEY, GEMINI_API_KEY (or ANTHROPIC_API_KEY), BRIGHTDATA_API_TOKEN
go mod tidy
go run ./cmd/ledgerlens

# 2) Frontend (Next.js)
cd apps/web
cp .env.local.example .env.local
pnpm install
pnpm dev
```

Open [http://localhost:3000/](http://localhost:3000/).

### Regenerate TypeScript types from Go schemas

```bash
go install github.com/gzuidhof/tygo@latest
make schemas
```

---

## Documentation

| Doc | Purpose |
|---|---|
| `Docs/Bright-Data-winning-strategy.md` | Source-of-truth strategy doc (v2.3) |
| `Docs/BRIGHTDATA_INTEGRATION.md` | Per-product role + code pointer + sample receipts |
| `Docs/GEM2_AUDIT_MODEL.md` | L0–L3 spec + canonical EEF + "Speculative" UI mapping |
| `Docs/X402_FLOW.md` | Simulation state machine + Settler interface + transparency rationale |
| `Docs/DEMO_SCRIPT.md` | 5-min demo storyboard (Case A blocked + Case B approved) |
| `Docs/JUDGE_GUIDE.md` | 5-bullet "look here first" for hackathon judges |
| `Docs/LEGAL.md` | Public-only scope · Bright Data AUP · simulation-only posture |
| `Docs/THIRD_PARTY_NOTICES.md` | MIT for own code · CC-BY-4.0 attribution for TPMN-PSL excerpts |

---

## License

MIT. See [`LICENSE`](./LICENSE) and [`Docs/THIRD_PARTY_NOTICES.md`](./Docs/THIRD_PARTY_NOTICES.md).

The deployed GEM² Truth-Filter SaaS at `gem2-tpmn-checker.fly.dev` (L1/L2 audit-gate upstream) is proprietary GEM².AI infrastructure that LedgerLens calls as a service.

---

*Built with GEM² · TPMN-PSL · Bright Data · Coinbase x402 (simulation) · Go · Next.js.*
