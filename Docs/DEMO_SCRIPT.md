# Demo Script — 5 minutes

> ⚠ **SKELETON (Unit 1).** Filled in Unit 6 (scenarios + video shoot).
> Strategy doc §10 is the canonical source until this skeleton is replaced.

## Opening — 30 sec

> "Autonomous agents are about to spend trillions of dollars. Today they hallucinate. Who guards the wallet?"

"LedgerLens is the **GEM² Trust Gate** between an agent and its wallet. It uses Bright Data to verify seller claims against live public web evidence, then GEM² audits the claims, and only then is payment authorized. Slogan: **No grounded claim, no payment.**"

> "For demo safety, we do not connect a personal Coinbase account or expose private keys. Instead, we simulate the x402 payment lifecycle. We simulate settlement, but the trust gate is real."

## Case A — BLOCKED — 90 sec

Show: seller claims "verified real-time pricing from the official vendor site."
Action: Bright Data fetches the source page. UI shows the evidence panel.
Action: GEM² L1 P-check fires. UI surfaces `[EVIDENCE-1] outdated snapshot`, `[SPT-S→T]` badge, `[EEF-⊬]` "Speculative" tag, `[RULE-1] FAIL`.
Action: L3 BLOCKS the payment.

**The wow moment:** the red banner "**Payment blocked: seller claim not grounded. Evidence is 6 months stale; cannot certify 'real-time'.**"

## Case B — APPROVED — 90 sec

Show: seller claims "Live NYSE+NASDAQ feed, 1-second freshness, $0.001/query."
Action: Bright Data Browser API verifies the live dashboard.
Action: L1 ALLOW (score 94). L2 SUCCESS (score 96).
Action: L3 APPROVED.
Action: SimulatedSettler emits `sim_x402_<id>` receipt; UI shows SIMULATION MODE pill + settlement transcript.

## Close — 30 sec

> "LedgerLens makes autonomous agent payments safe by forcing every web-data purchase through live evidence verification and a GEM² Trust Gate before x402 settlement."

> "The Trust Gate is implemented via our production audit-gate API at `gem2-tpmn-checker.fly.dev`. Real x402 settlement is one interface swap away."

## Buffer / Q&A — 1 min

- Bright Data products in use: SERP + Unlocker + Browser + MCP (+ optional Scraper).
- Architecture: one Go binary + Next.js UI.
- Why simulation: public demo safety, focus on the gate.

## Recording targets

- Total: ≤ 4 min 30 sec (buffer for upload).
- Captioned.
- Single-take or tight edit.
- Public live URL must be alive during recording.
