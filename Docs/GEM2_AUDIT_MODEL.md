# GEM² Audit Model

How LedgerLens uses GEM² across the audit stack.

> ⚠ **SKELETON (Unit 1).** Filled in Unit 3 (GEM² Trust Gate integration).
> Strategy doc §5 is the canonical source until this skeleton is replaced.

## Two complementary framings

LedgerLens uses **both**:

1. **Architectural stack (high-level)** — four layers describing the audit pipeline as a whole.
2. **Per-contract gate pair (production canonical)** — L1 P-check + L2 O-check, as deployed at `gem2-tpmn-checker.fly.dev`.

### Architectural stack

| Layer | Purpose | Rule |
|---|---|---|
| **Evidence** | Forensic acquisition via Bright Data | If there is no source receipt, no claim may proceed. |
| **Memory** | Entity / offer / seller / buyer / source memory | If entity identity or source context is ambiguous, downgrade the claim. |
| **Verification** | Claim scoring + overclaim detection — *implemented via the deployed audit-gate API* | Unsupported or speculative claims cannot trigger payment approval. |
| **Release** | Composite policy gate — local Go | x402 settlement is allowed only after composite L1∧L2 + policy release. |

### Per-contract gate pair

```
For every (BuyerRequest, SellerOffer) candidate pair:
  ① L1 P-check  → POST gem2-tpmn-checker.fly.dev/api/audit-gate/p-check
  ② Local F     → compose draft DecisionPacket
  ③ L2 O-check  → POST gem2-tpmn-checker.fly.dev/api/audit-gate/o-check
  ④ L3 Release  → composite verdict; APPROVED ⟹ simulated settlement
```

## Canonical EEF — 4 tags

API + audit bundle use the canonical 4 tags. "Speculative" is a UI-only label for `extrapolated` with no stated basis.

| Tag | API value | UI label | Meaning |
|---|---|---|---|
| ⊢ | `grounded` | "Grounded" | Confirmed by direct evidence |
| ⊨ | `inferred` | "Inferred" | Derived from grounded with visible chain |
| ⊬ | `extrapolated` (basis present) | "Extrapolated" | Beyond evidence; basis stated |
| ⊬ | `extrapolated` (basis empty) | **"Speculative"** | Beyond evidence; **no** basis stated |
| ⊥ | `unknown` | "Unknown" | Knowledge gap |

## SPT guardrails (production canonical)

| Class | Code value | Meaning | LedgerLens demo trigger |
|---|---|---|---|
| State→Trait | `S->T` | Contextual finding presented as permanent | "99.9% uptime" claim from one week of data |
| Local→Global | `L->G` | One case generalized to all | "Universally accurate pricing" from US-only coverage |
| Increment-as-Mass | `delta_e->int_de` | Sparse data presented as established trend | "+12% MoM growth proves the trend" from two data points |

## References

- `Gem-squared-AI/gem2-TPMN-checker/AUDIT_GATE_API.md` v1.1
- `Docs/Bright-Data-winning-strategy.md` §5
- TPMN-PSL grammar primer (~/.claude/skills/set-persona/references/tpmn-grammar-primer.md)
