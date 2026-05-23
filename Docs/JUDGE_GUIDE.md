# Judge Guide — Where to Look First

> ⚠ **SKELETON (Unit 1).** Filled in Unit 7 (submission pack).

## The 5-bullet quick start

1. **Open the live app** — [LIVE URL pending Unit 5 deploy] — click **▶ Run Case A** to see the **BLOCKED** payment, then **▶ Run Case B** for the **APPROVED** payment.
2. **Look at the audit panel** on the right side of each case. The `[EEF-…]` tags + `[SPT-…]` badges are the originality moat — visible epistemic verification BEFORE money moves.
3. **Note the `SIMULATION MODE` pill** near every settlement event. We chose to keep personal Coinbase accounts and private keys out of a public demo; the `Settler` interface is the post-hackathon swap point for real x402.
4. **Browse `Docs/BRIGHTDATA_INTEGRATION.md`** for the per-product role table — LedgerLens uses 4 Bright Data products in distinct roles (SERP + Unlocker + Browser + MCP), not one MCP wrapper.
5. **Open `artifacts/audit_bundles/`** for sample exported decision packets — every payment (approved or blocked) produces a regulator-replay-ready bundle with hash-chained evidence and upstream `gem2-tpmn-checker` `result_id`s.

## Judging criteria — where to score us

| Criterion | Look here |
|---|---|
| **Application of Tech** | `Docs/BRIGHTDATA_INTEGRATION.md` (4 products) + `Docs/GEM2_AUDIT_MODEL.md` (production audit-gate API) + `Docs/X402_FLOW.md` (x402 lifecycle simulation) |
| **Presentation** | Live demo + 5-min video + the slogan "No grounded claim, no payment." |
| **Business Value** | CFOs can't approve agentic spend because models hallucinate. LedgerLens unlocks the spend by gating it on grounded evidence. |
| **Originality** | The visible Trust Gate before payment, the SPT overclaim guards, the simulation-mode transparency posture |

## Source-of-truth doc

`Docs/Bright-Data-winning-strategy.md` (v2.3) — full architecture, multi-track positioning, and decision history.
