# GEM² Audit Model

How LedgerLens uses GEM² across the audit stack. Implementation lives in `internal/trustgate/`; the upstream gem2-tpmn-checker API spec is the GEM² audit-gate API spec (proprietary internal documentation) v1.1.

## Two complementary framings

### Architectural stack (high-level)

| Layer | Purpose | Rule | Where it lives |
|---|---|---|---|
| **Evidence** | Forensic acquisition via Bright Data | If there is no source receipt, no claim may proceed. | `internal/brightdata/` (Unit 2) + `internal/trustgate/evidence/` (this unit's wrapping) |
| **Memory** | Entity / offer / seller / buyer / source memory | If entity identity or source context is ambiguous, downgrade the claim. | `internal/trustgate/memory/` (SQLite `gate_decisions` table — hash-only mirror) |
| **Verification** | Claim scoring + overclaim detection — *implemented via the deployed audit-gate API* | Unsupported or speculative claims cannot trigger payment approval. | `internal/trustgate/auditgate/` (Go client to `gem2-tpmn-checker.fly.dev`) |
| **Release** | Composite policy gate before settlement — local Go | x402 settlement is allowed only after composite L1∧L2 + policy release. | `internal/trustgate/release/` (Unit 4) + `internal/paymentgate/` (Unit 4) |

### Per-contract gate pair (production canonical)

```
For every (BuyerRequest, SellerOffer) candidate pair:
  ① L1 P-check  → POST gem2-tpmn-checker.fly.dev/api/audit-gate/p-check
  ② Local F     → compose draft DecisionPacket
  ③ L2 O-check  → POST gem2-tpmn-checker.fly.dev/api/audit-gate/o-check
  ④ L3 Release  → composite verdict; APPROVED ⟹ simulated settlement
```

## Canonical EEF — 4 tags

| Tag | API value | UI label | Meaning |
|---|---|---|---|
| ⊢ | `grounded` | "Grounded" | Confirmed by direct evidence |
| ⊨ | `inferred` | "Inferred" | Derived from grounded with visible chain |
| ⊬ | `extrapolated` (basis present) | "Extrapolated" | Beyond evidence; basis stated |
| ⊬ | `extrapolated` (basis empty) | **"Speculative"** | Beyond evidence; **no** basis stated |
| ⊥ | `unknown` | "Unknown" | Knowledge gap |

## SPT guardrails (production canonical)

| Class | Code value | Meaning | Demo trigger |
|---|---|---|---|
| State→Trait | `S→T` | Contextual finding presented as permanent | "99.9% uptime" claim from one week of data |
| Local→Global | `L→G` | One case generalized to all | "Universally accurate pricing" from US-only coverage |
| Increment-as-Mass | `Δe→∫de` | Sparse data presented as established trend | "+12% MoM growth proves the trend" from two data points |

## Reason parsing — how the gate's flat strings become structured claims

The audit gate emits a flat `reasons: []string` where each entry is a bracketed tag:

```
[TYPE] I lacks required fields: offerId, sellerId, sourceUrl missing
[RULE-1] claim must be supported by provided evidence — PASS: NYSE+NASDAQ coverage confirmed by vendor status page
[RULE-2] price must be within buyer.maxSpendUSDC — PASS: $0.001 is well below $0.01 cap
[RULE-3] no Δe→∫de overclaims — PASS: no trend claims made, specific pricing stated
[SPT-S→T] <only if a violation found>
[EEF-⊬] <only if extrapolation flagged>
[EVIDENCE-1] vendor status page — used as ⊢ grounded fact for RULE-1
[DIM-grounding] 82/100 — claim aligns with provided evidence
[DIM-evidence] 90/100 — direct status page provided
[DIM-logical] 88/100 — pricing math checks out
```

[`internal/trustgate/auditgate/reasons.go`](../internal/trustgate/auditgate/reasons.go) parses these with six regex patterns into `ParsedReasons`, then `ToClaimAssessments` derives one `schemas.ClaimAssessment` per `[RULE-N]`:

| Rule verdict | Evidence correlated? | → ClaimStatus | Basis |
|---|---|---|---|
| PASS | ≥1 [EVIDENCE-N] mentions the rule | `grounded` (⊢) | (empty) |
| PASS | none | `inferred` (⊨) | (empty) |
| FAIL | any | `extrapolated` (⊬) | the "FAIL: \<why\>" text |
| FAIL | [EEF-⊥] flag + "unknown"/"no evidence" in reason | `unknown` (⊥) | (empty) |

SPT codes from `[SPT-X]` and Confidence (from response Score / 100) are attached to every claim.

## Local audit trail — `gate_decisions` table

[`internal/trustgate/memory/store.go`](../internal/trustgate/memory/store.go) mirrors the upstream hash-only schema (per AUDIT_GATE_API.md §"Audit Trail"):

```sql
CREATE TABLE gate_decisions (
  id            TEXT PRIMARY KEY,        -- matches upstream meta.result_id
  gate_type     TEXT NOT NULL,           -- 'P_GATE' or 'O_GATE'
  verdict       TEXT NOT NULL,
  score         INTEGER NOT NULL,
  threshold     INTEGER NOT NULL,
  input_hash    TEXT NOT NULL,           -- SHA-256 of i (or o)
  p_hash        TEXT NOT NULL,
  evidence_hash TEXT,                    -- nullable when no evidence
  provider      TEXT NOT NULL,
  duration_ms   INTEGER NOT NULL,
  created_at    TEXT NOT NULL,
  reasons_json  TEXT                     -- LedgerLens local extension
);
```

Raw `i` / `o` / `p` / `evidence` content is **never persisted** — only SHA-256 hashes. The `reasons_json` extension keeps the full reasons array locally so claim assessments can be reconstructed without re-calling the gate.

## REPLAY mode — demo stability fallback

[`internal/trustgate/auditgate/replay.go`](../internal/trustgate/auditgate/replay.go) — file-backed cache at `artifacts/demo_cases/<case>/{p_check,o_check}.json`. When `gem2-tpmn-checker.fly.dev` is unreachable AND the `Client` is wired with a `ReplayStore` AND a `replayCaseID` is passed, the client returns the cached response.

```
Demo seeded scenarios:
  artifacts/demo_cases/case_a/p_check.json  ← BLOCKED scenario
  artifacts/demo_cases/case_b/p_check.json  ← APPROVED scenario
  artifacts/demo_cases/unit3_smoke/p_check.json  ← test fixture (this unit)
  artifacts/demo_cases/unit3_smoke/o_check.json
```

UI exposes a **LIVE / REPLAY** toggle (Unit 5). When toggle is LIVE: hit the API; on failure, fall back to REPLAY with a visible "REPLAY MODE" pill.

## Re-running the audit tests

```bash
# Offline (no env required)
go test -v -count=1 -short ./internal/trustgate/...

# Live round-trip (requires .env with GEM2_API_KEY + ANTHROPIC_API_KEY)
set -a && source .env && set +a
go test -v -count=1 -timeout 240s ./internal/trustgate/auditgate/...
```

## References

- GEM² audit-gate API spec v1.1 (proprietary; available via GEM².AI)
- `Hackathon/TechEx/console/audit_gate_client.go` (reference Go integration, MIT)
- `Docs/Bright-Data-winning-strategy.md` §5 (high-level architecture)
