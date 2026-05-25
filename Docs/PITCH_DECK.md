---
marp: true
theme: gaia
class: invert
paginate: true
size: 16:9
header: 'LedgerLens × Bright Data — No grounded claim, no payment.'
footer: 'https://ledgerlens.gemsquared.ai · github.com/gem-squared/ledgerlens · MIT'
---

<!-- _class: lead invert -->

# LedgerLens × Bright Data

## The trust layer between agent intent and x402 settlement.

- Autonomous agents are about to spend USD billions via **x402** and **MCP-pay**
- The web they shop on is hostile: scraped, gated, faked, stale
- Today's default is **pay-first, verify-never** — that's the bug LedgerLens fixes

> Built for the Bright Data "Web Data UNLOCKED" Hackathon · May 25 – 31, 2026 · SF Web Data Loft

---

# No grounded claim, no payment.

- **Buyer agent** extracts intent + spend policy from the judge's natural-language request
- **Bright Data** fetches public-web evidence; a seller agent proposes an offer
- **GEM² Trust Gate** releases (or blocks) x402-shaped settlement based on whether the seller's claims actually ground

> We simulate settlement — the trust gate is real.

---

# Six contracts. One direction.

- **JUDGE → BUYER AGENT → BRIGHT DATA → SELLER OFFER → GEM² GATE → X402 SETTLE → AUDIT BUNDLE**
- Every box has a typed CONTRACT (A → B | P) — no silent state, no implicit trust between agents
- Every decision (APPROVED or BLOCKED) exports a **hash-chained audit bundle**, replayable by any regulator

> Money only moves after the last green check.

---

# Bright Data is the evidence substrate.

- **SERP API** discovers candidate sellers from live web search — zone `gem2_serp_api1`
- **Web Unlocker + Browser API** fetch and render evidence pages public web hides — `gem2_web_unlocker1`, `gem2_scraping_browser1`
- **MCP server** (`@brightdata/mcp`, 5 tools) gives the buyer agent a natural-language tool surface

> Four Bright Data products in distinct roles — not one MCP wrapper.

---

# GEM² Trust Gate — two-stage formal audit.

- **L1 P-check** — proves the buyer's preconditions hold against evidence (ALLOW / DENY)
- **L2 O-check** — proves the seller's claims are actually grounded (SUCCESS / FAILURE)
- Canonical EEF tags on every claim: **⊢ grounded · ⊨ inferred · ⊬ extrapolated · ⊥ unknown**

> Powered by `gem2-tpmn-checker.fly.dev` — independent, replayable, ~10–20s round-trip per deal.

---

# x402 simulation — protocol-correct, zero counterparty risk.

- `SimulatedSettler` emits x402-shaped receipts: `settlementId` · `amountUSDC` · `mode=simulation`
- Every receipt provably safe: `real_transaction=false`, `realFundsUsed=false`, `privateKeysUsed=false`
- Drop-in to the Coinbase x402 Go SDK post-hackathon — same `Settler` interface, zero refactor

> The hackathon ships the trust gate, not the wallet. The hardest part is built.

---

# Live, working, audited.

- **34 deals audited end-to-end** · **12 approved · 22 blocked** · **87 Bright Data evidence receipts** · ~19s avg verification latency
- Three modes: **LIVE** (real Bright Data + real GEM² audit, 20-45s) · **REPLAY** Case A (BLOCKED — stale evidence) + Case B (APPROVED — fresh evidence) · **VIEW** any historical bundle
- One Go binary + embedded Next.js + atomic `make deploy` with auto-rollback — `https://ledgerlens.gemsquared.ai/`

> Click any row in Recent Audit Samples to inspect the receipts the agents acted on.

---

# Same gate. Three judging lenses.

- **T-Agent (primary):** autonomous buyer/seller settling via x402 simulation, spend caps enforced at L3, approval checkpoint visible
- **Track 3 — Risk Management (cross-claim):** third-party risk on sellers, regulator-replayable bundles, open-web evidence no SIEM watches — EEF + SPT tags are the structured intelligence
- **T-Commerce:** sellers price web-data offers in USDC-demo, buyers verify freshness + coverage before paying — micropayments ($0.001/query) gated by truth

> GEM² Trust Gate is the substrate. LedgerLens is the first product shape on top.

---

<!-- _class: lead invert -->

# No grounded claim, no payment.

- **LIVE:** https://ledgerlens.gemsquared.ai/ — click ▸ Run Autonomous Deal
- **SOURCE:** github.com/gem-squared/ledgerlens (MIT) — single Go binary, embedded Next.js
- **TEAM:** gem-squared · david@gineers.ai

> Fast agents are dangerous if they spend before verification. **LedgerLens deliberately waits.**
