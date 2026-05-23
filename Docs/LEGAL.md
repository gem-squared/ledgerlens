# LEGAL.md

## Scope

LedgerLens is a hackathon prototype built for the **Bright Data "Web Data UNLOCKED" Hackathon** (lablab.ai, 2026-05-25 → 2026-05-31).

## Settlement posture — SIMULATION ONLY

The hackathon repo ships only `SimulatedSettler`. This means:

- ⊢ **No private keys** are stored in the repository, environment, browser, or runtime memory.
- ⊢ **No Coinbase or other custodial accounts** are connected.
- ⊢ **No real funds** (mainnet or testnet) are transacted by the hackathon build.
- ⊢ **No on-chain transactions** (mainnet or Base Sepolia) are produced.
- ⊢ All settlement records carry `mode: "simulation"` and `real_transaction: false` for full transparency.

A real x402 settlement implementation is a post-hackathon swap behind the `Settler` interface and is **not** part of this submission.

## Data collection posture

LedgerLens uses Bright Data products exclusively for **public web sources**. Per the Bright Data Acceptable Use Policy:

- ⊢ No login-required collection.
- ⊢ No nonpublic / private data.
- ⊢ No exploit verification, credential use, or active probing.
- ⊢ robots.txt-aware fetch policy enforced in `internal/brightdata/aup.go`.
- ⊢ Source-tier metadata captured per fetch for the audit bundle.

## Personal data minimization

- ⊢ The L2 audit panel scrubs or masks unnecessary personal data from analyst-visible summaries before display.
- ⊢ The hash-only audit-trail design of `gem2-tpmn-checker` (upstream L1/L2 service) means raw input content is never persisted on the upstream side; only SHA-256 hashes are stored.

## Compliance posture (thin surface, honest)

LedgerLens does not perform regulatory compliance (KYC/AML, GDPR, SOX, etc.). It performs **epistemic** compliance: verifying that a claim is grounded in evidence before a payment is authorized. The thin compliance surface that ships in the MVP:

| Surface | Implementation |
|---|---|
| Spending cap | `PaymentPolicy.spendCap` enforced at L3 |
| Public-only collection | `internal/brightdata/policy.go` |
| Immutable audit log | `artifacts/audit_bundles/` (hash-chained, append-only) |
| AUP-aware source policy | `internal/brightdata/aup.go` |
| Decision packet | Every payment produces a `DecisionPacket` (approved or blocked) |

## Licensing

- LedgerLens application code: **MIT** (see `LICENSE`).
- TPMN-PSL spec excerpts and TPMN Skill standard text (if quoted): **CC-BY-4.0**. See `Docs/THIRD_PARTY_NOTICES.md` for attribution.
- The deployed `gem2-tpmn-checker.fly.dev` SaaS (called by LedgerLens) is **proprietary GEM².AI** infrastructure.

## Authors

- Inseok "David" Seo (`david@gineers.ai`) — GEM².AI.
