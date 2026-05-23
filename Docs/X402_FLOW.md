# x402 Flow (Simulation Mode)

How LedgerLens models the x402 payment lifecycle without any chain dependency.

> ⚠ **SKELETON (Unit 1).** Filled in Unit 4 (x402 simulation + L3 gate).
> Strategy doc §8 is the canonical source until this skeleton is replaced.

## The lock

**Hackathon repo ships SIMULATION ONLY.** No private keys. No Coinbase account. No Base Sepolia dependency. No `tx_hash`. No block-explorer link.

## State machine

```
PAYMENT_REQUIRED
    │
    ▼
PENDING_VERIFICATION  ── Evidence + Memory + L1 P-check + F + L2 O-check + L3 Release
    │
    ├─ APPROVED_BY_TRUST_GATE  ─→  SIMULATED_SETTLED   (settlement_id: sim_x402_<id>)
    ├─ BLOCKED_BY_TRUST_GATE   ─→  no settlement       (settlement_id: null)
    └─ ESCALATED_TO_HUMAN      ─→  held                (operator decision)
```

## Settlement receipt shape

### APPROVED

```json
{
  "settlementId": "sim_x402_a3f17c92",
  "decisionId":   "<uuid>",
  "mode":         "simulation",
  "network":      "demo-local",
  "asset":        "USDC-demo",
  "status":       "SIMULATED_SETTLED",
  "amountUSDC":   0.001,
  "reason":       "L3 Trust Gate approved grounded claim",
  "realTransaction":  false,
  "privateKeysUsed":  false,
  "realFundsUsed":    false,
  "ts": "2026-05-31T17:42:11Z"
}
```

### BLOCKED

```json
{
  "settlementId": null,
  "decisionId":   "<uuid>",
  "mode":         "simulation",
  "status":       "BLOCKED_BY_TRUST_GATE",
  "amountUSDC":   0,
  "reason":       "<L1/L2 reason chain>",
  "realTransaction": false
}
```

## Why simulation

- **Public demo safety.** No personal Coinbase account, no private key in repo or browser.
- **Demo stability.** No external chain to fail mid-presentation.
- **Focus.** The differentiator is the Trust Gate, not the chain. Settlement is the *consequence* of verification.
- **Pitch line.** *"We simulate settlement, but the trust gate is real."*

## Post-hackathon — real x402

The `Settler` interface in `internal/schemas/types.go` is the swap point:

```go
type Settler interface {
    Settle(ctx context.Context, decision DecisionPacket) (SimulatedSettlement, error)
}
```

Replace `SimulatedSettler` with a `CoinbaseX402Settler` that wraps the official Coinbase Go SDK + CDP facilitator + Base Sepolia (later: mainnet under proper key custody). The Trust Gate code does not change.

## References

- `Docs/Bright-Data-winning-strategy.md` §8 (canonical flow)
- Coinbase x402 official docs: https://docs.cdp.coinbase.com/x402/welcome (post-hackathon reference)
- Coinbase x402 Go reference: https://github.com/coinbase/x402/tree/main/go (post-hackathon reference)
