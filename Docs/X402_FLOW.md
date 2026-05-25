# x402 Flow (Simulation Mode)

How LedgerLens models the x402 payment lifecycle without any chain dependency. Filled in Unit 4.

**Status (2026-05-23):** end-to-end pipeline LIVE; Case A (BLOCKED) and Case B (APPROVED) both green.

## The lock — v2.2

**Hackathon repo ships SIMULATION ONLY.** No private keys. No Coinbase account. No Base Sepolia dependency. No `tx_hash`. No block-explorer link.

| Invariant | Value | Where enforced |
|---|---|---|
| `mode` | `simulation` | hardcoded in `SimulatedSettler` |
| `network` | `demo-local` | hardcoded |
| `asset` | `USDC-demo` | hardcoded |
| `real_transaction` | `false` | every settlement record |
| `private_keys_used` | `false` | every settlement record |
| `real_funds_used` | `false` | every settlement record |
| Coinbase Go SDK in `go.mod` | **absent** | `go.mod` does not depend on `github.com/coinbase/x402/go` |

## State machine

```
PAYMENT_REQUIRED
    │
    ▼
PENDING_VERIFICATION  ── Evidence + Memory + L1 P-check + F + L2 O-check + L3 Release
    │
    ├─ APPROVED_BY_TRUST_GATE  ─→  SIMULATED_SETTLED   (settlement_id: sim_x402_<hex8>)
    ├─ BLOCKED_BY_TRUST_GATE   ─→  no settlement       (settlement_id: null)
    └─ ESCALATED_TO_HUMAN      ─→  held                (operator decision)
```

The state machine + transitions live in [`internal/paymentgate/simulated_settler.go`](../internal/paymentgate/simulated_settler.go). The Settler interface in [`internal/schemas/types.go`](../internal/schemas/types.go) is the swap point.

## L3 composite verdict

The release gate ([`internal/trustgate/release/gate.go`](../internal/trustgate/release/gate.go)) combines L1 + L2 + policy into a single verdict:

```
APPROVED_BY_TRUST_GATE iff
    L1.verdict = ALLOW
  ∧ L1.score ≥ T_L1                  (default 70)
  ∧ L2.verdict = SUCCESS
  ∧ L2.score ≥ T_L2                  (default 75)
  ∧ proposed_spend ≤ policy.spendCap

BLOCKED_BY_TRUST_GATE iff
    L1.verdict = DENY                (or L1 unreachable + no replay)
  ∨ L1.score < T_L1
  ∨ L2.verdict = FAILURE
  ∨ L2.score < T_L2
  ∨ proposed_spend > policy.spendCap
```

The MVP collapses to APPROVED / BLOCKED binary; ESCALATED_TO_HUMAN intermediate band is reserved for Unit 5's UI surface.

## Orchestrator

[`internal/paymentgate/orchestrator.go`](../internal/paymentgate/orchestrator.go) composes the full pipeline:

```
Orchestrator.Run(buyer, offer, []EvidenceReceipt) →
    ① evidence.WrapReceipts                      → []string chunks
    ② auditgate.PCheck                           → L1 P-check response
    ③ F: compose draft DecisionPacket            → with ClaimAssessment[] from L1 reasons
    ④ auditgate.OCheck (only if L1 ALLOW + ≥T)   → L2 O-check response
    ⑤ release.Compose(L1, L2, policy, spend)     → L3 Decision
    ⑥ Settler.Settle(decision)                   → SimulatedSettlement
    ⑦ BundleStore.Write + memory.Store.Insert    → AuditBundle + gate_decisions rows
    ↓
returns (DecisionPacket, SimulatedSettlement, bundlePath, error)
```

## Settlement receipt shapes

### APPROVED

```json
{
  "settlementId":     "sim_x402_631afb862b5fd168",
  "decisionId":       "dp_<hex8>",
  "mode":             "simulation",
  "network":          "demo-local",
  "asset":            "USDC-demo",
  "status":           "SIMULATED_SETTLED",
  "amountUSDC":       0.001,
  "reason":           "L3 Trust Gate approved grounded claim",
  "realTransaction":  false,
  "privateKeysUsed":  false,
  "realFundsUsed":    false,
  "ts":               "2026-05-23T11:35:00Z"
}
```

### BLOCKED

```json
{
  "decisionId":       "dp_<hex8>",
  "mode":             "simulation",
  "network":          "demo-local",
  "asset":            "USDC-demo",
  "status":           "BLOCKED_BY_TRUST_GATE",
  "amountUSDC":       0,
  "reason":           "L1 P-check verdict=DENY score=15 — claim not grounded",
  "realTransaction":  false,
  "privateKeysUsed":  false,
  "realFundsUsed":    false,
  "ts":               "2026-05-23T11:35:00Z"
}
```

Note: `settlementId` is omitted (null) when blocked.

## Audit bundle

Every decision writes one JSON bundle to `artifacts/audit_bundles/<decisionId>.json` via [`internal/paymentgate/audit_bundle.go`](../internal/paymentgate/audit_bundle.go).

```json
{
  "bundleId":         "ab_<hex8>",
  "decisionId":       "dp_<hex8>",
  "buyerRequest":     { ... },
  "sellerOffer":      { ... },
  "evidenceReceipts": [ ... ],
  "evidenceHash":     "sha256:...",
  "l1":               { verdict, score, reasons, meta },
  "l2":               { verdict, score, reasons, meta },     // omitted if L1 denied
  "claimAssessments": [ ... ],
  "decision":         { ... },
  "settlement":       { ... },
  "timestamps":       { "startedAt": "...", "finishedAt": "..." }
}
```

The bundle is regulator-replay-ready: contains both upstream `result_id`s, hash-chained evidence, and the full claim trace. Reasons preserved for offline reconstruction.

## Why simulation

- **Public demo safety.** No personal Coinbase account, no private key in repo or browser.
- **Demo stability.** No external chain to fail mid-presentation.
- **Focus.** The differentiator is the Trust Gate, not the chain. Settlement is the *consequence* of verification.
- **Pitch line.** *"We simulate settlement, but the trust gate is real."*

## Post-hackathon — real x402

The `Settler` interface in [`internal/schemas/types.go`](../internal/schemas/types.go) is the swap point:

```go
type Settler interface {
    Settle(ctx context.Context, decision DecisionPacket) (SimulatedSettlement, error)
}
```

Replace `SimulatedSettler` with a `CoinbaseX402Settler` that wraps the official Coinbase Go SDK + CDP facilitator + Base Sepolia (later: mainnet under proper key custody). The Trust Gate code does not change.

## Re-running the e2e

```bash
# Offline unit tests (no env required)
go test -v -count=1 ./internal/paymentgate/... -run 'TestSimulatedSettler|TestBundleStore'
go test -v -count=1 ./internal/trustgate/release/...

# Live end-to-end (requires .env with GEM2_API_KEY + ANTHROPIC_API_KEY)
set -a && source .env && set +a
go test -v -count=1 -timeout 240s ./internal/paymentgate/...
```

Live cost per full e2e run: ~$0.066 (3 audit-gate calls × ~$0.022, claude-sonnet-4).

## References

- `Docs/Bright-Data-winning-strategy.md` §8 (canonical flow, simulation rationale)
- `Docs/GEM2_AUDIT_MODEL.md` (audit layer details)
- GEM² audit-gate API spec (proprietary; available from GEM².AI)
- Coinbase x402 docs (post-hackathon reference): https://docs.cdp.coinbase.com/x402/welcome
