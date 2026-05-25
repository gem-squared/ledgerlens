# LedgerLens — Bright Data "Web Data UNLOCKED" Winning Strategy

> **No grounded claim, no payment.**
>
> **Product:** LedgerLens — an **x402-native Agent-to-Agent Payments product** with a **Bright Data-powered web evidence layer** and a **GEM² Trust Gate** before settlement.
>
> **Event:** Web Data UNLOCKED Hackathon — Bright Data + lablab.ai
> **Dates:** 2026-05-25 → 2026-05-31 (online build May 25–30 · hybrid day May 30 · demos & awards May 31)
> **Today:** 2026-05-23 — T-2 to kickoff, T+8 to demos
> **Venue:** Hybrid — online anywhere · onsite Bright Data SF "Web Data Loft", 625 2nd St, San Francisco. **Plan: onsite.**
> **Grand prize:** $5,000 + fast-track to Bright Data AI Startup Program (up to ~$20,000 in credits)
> **Mandatory:** demonstrable use of ≥1 Bright Data product
>
> **Owner:** David (Gineers / GEM²)
> **Status:** v2.3 · last updated 2026-05-23 · supersedes v2.2
> **Authoring trace:** Alchy → Kritik (Go stack) → Alchy (v2.0) → Kritik (audit-gate API integration) → Alchy (v2.1 — §5 dual framing) → David (settlement = simulation) → Alchy (v2.2) → Kritik (v2.1/v2.2 residue audit + gem2-checker fallback risk) → Alchy (v2.3 — consistency patch: all chain residues purged from execution-current claims; gem2-tpmn-checker REPLAY-mode fallback added to §12)

---

## 0. TL;DR — the win plan in nine lines

1. ⊢ **Product positioning:** LedgerLens is an x402-native Agent-to-Agent Payments product with a Bright Data web-evidence layer and a GEM² Trust Gate before settlement.
2. ⊢ **Slogan:** **No grounded claim, no payment.**
3. ⊢ **Primary track:** Agent-to-Agent Payments / Autonomous Agents (two agents settling autonomously). Claim Web Data Commerce + General Web Data as multi-track.
4. ⊢ **Stack lock:** **Go-first** — one Go binary owns the Trust Gate, audit bundle, decision packet, x402 settlement state, and Bright Data wrappers. Next.js owns the demo UI. TS shim slots reserved but probably unused.
5. ⊢ **Originality moat:** the GEM² Trust Gate — visible epistemic verification BEFORE money moves. No other team will ship this. **L1 P-check + L2 O-check are integrated via the deployed `gem2-tpmn-checker.fly.dev` audit-gate API (GEM².AI-operated, production-tested via TechEx); we do not rebuild it.**
6. ⊢ **Bright Data depth:** SERP + Web Unlocker + Browser API + MCP Server (Web Scraper API optional). Four products in distinct roles, not one MCP wrapper.
7. ⊢ **x402:** **protocol SIMULATION mode** — we model the HTTP 402 lifecycle locally (PAYMENT_REQUIRED → PENDING_VERIFICATION → APPROVED/BLOCKED → SIMULATED_SETTLED). No real chain, no Coinbase account, no private keys in the public demo. Settlement receipt is JSON tagged `mode: "simulation"`. Real x402 settlement = post-hackathon swap behind a `Settler` interface.
8. ⊢ **Demo spine (≤5 min):** seller-agent posts offer → buyer-agent requests data → Bright Data fetches evidence → GEM² L1+L2 audit-gate scores → L3 APPROVES or **BLOCKS** payment → simulated settlement receipt → audit bundle exports.
9. ⊢ **Build order:** T-2 contracts + repo, T+0 kickoff, T+1 Bright Data ingest, T+2 GEM² L0–L2, T+3 x402 + L3 gate, T+4 UI, T+5 polish + video, T+6 onsite + submit.

---

## 1. What changed from v1.0 (and from the prior research report)

| Aspect | v1.0 / report | v2.x (this doc) | Reason |
|---|---|---|---|
| Track framing | "Track 3: Security & Compliance" | **x402 Agent-to-Agent Payments + Web Data Commerce + GEM² Trust Gate** | Hackathon theme is agentic commerce, not security/compliance |
| Product third-vector | "Compliance Gate" | **Trust Gate / Epistemic Settlement Gate** | "Compliance" signals KYC/AML/GDPR — false expectation. Trust Gate names what we actually do. |
| Stack | ambiguous (Python residue from report) | **Go-first** (with TS shim slots) | Ground test 2026-05-23 confirms x402 Go is tier-1; MCP-Go is production; mcp-go-x402 exists |
| Claim taxonomy | 5-tag (added "speculative") | **Canonical EEF — 4 tags: ⊢ ⊨ ⊬ ⊥**. "Speculative" is a UI label for ⊬-without-basis | Preserves universal EEF discipline; UI stays vivid |
| Architecture | "4 separate agents/services" | **One Go binary, contracts internally** (buyer / seller / verifier / gate as ROLES, not processes) | 7 deployment surfaces in 8 days = lost hours |
| Compliance | implicit | **Thin compliance surface explicit:** spending cap · public-only · immutable audit log · AUP-aware source policy | Keeps the word honest without overpromising |
| **L1 / L2 verification** (v2.1) | "build LLM-driven verifier from scratch" | **Integrate deployed `gem2-tpmn-checker.fly.dev` audit-gate API** (P-check + O-check, v1.1 with RAG evidence) | GEM².AI operates the platform; TechEx proved the integration path; eliminates the largest prompt-engineering risk in the 8-day arc. ⊢ References: the GEM² audit-gate API spec (proprietary internal documentation) + TechEx audit-gate Go client reference (MIT). |
| **Layer naming** (v2.1) | L0/L1/L2/L3 = Evidence/Memory/Verify/Release | **Dual framing**: Architectural stack (Evidence/Memory/Verification/Release) PLUS per-contract gate pair (L1 P-check + L2 O-check from production canonical) | Both framings are real; production API uses L1/L2 to mean the gate pair INSIDE the Verification layer. See §5. |
| **Lobster Trap** (v2.1) | not in scope | **Optional regex-only L0/L3 (T+2 stretch)** ported from TechEx — pure-Go DPI, no LLM cost | Cheap defense; answers prompt-injection-in-scraped-pages risk proactively in the demo |
| **Settlement** (v2.2) | "Coinbase Go SDK on Base Sepolia, real tx_hash" | **x402 protocol SIMULATION** — Go state machine, no chain, no private keys, no testnet funds | Public demo cannot expose personal accounts. Differentiator is the Trust Gate, not the chain. Settlement receipt tagged `mode: "simulation"`. Real x402 = post-hackathon swap behind `Settler` interface. Pitch line: *"We simulate settlement, but the trust gate is real."* |
| **Consistency patch** (v2.3) | residual `Base Sepolia` / `tx_hash` / `coinbase/x402/go` / `mcp-go-x402` mentions in execution-current claims | purged from §4.2 P, §4.3, §9 T+3, §9 scope-cut, §11 X402_FLOW; kept only in changelog / post-hackathon / ecosystem-context references. New gem2-tpmn-checker REPLAY-mode fallback added to §12. | Pre-build consistency audit by Kritik (David, 2026-05-23). Stops the build from inheriting stale assumptions. |

⊨ Carried over unchanged: GEM² L0/L1/L2/L3 layer architecture · TPMN contract discipline · EEF + SPT guardrails · Bright Data product-routing principle · MIT/CC-BY licensing care · submission deliverables list.

---

## 2. Verified hackathon ground (⊢ as of 2026-05-23)

```
Event ≜ [
  name:            "Web Data UNLOCKED",
  host:            Bright Data,
  platform:        lablab.ai,
  format:          hybrid (online + SF onsite),
  onsite_venue:    "Web Data Loft, 625 2nd St, SF" (Bright Data SF opening),
  start:           2026-05-25,
  build_phase:     2026-05-25 → 2026-05-30 (online),
  hybrid_day:      2026-05-30,
  demos_awards:    2026-05-31,
  credits:         $250 Bright Data API credits per participant (Day 1),
  mcp_free_tier:   5,000 requests / month,
  grand_prize:     $5,000 cash,
  bonus:           fast-track to Bright Data AI Startup Program (~$20K credits possible),
  mandatory:       demonstrable use of ≥1 Bright Data product,
  multi_track:     allowed (single submission may span multiple tracks)
]
```

### 2.1 Tracks (composite from observed listings)

| Track | Working title | Brief |
|---|---|---|
| **T-Agent** | Autonomous Agents | ⊢ Two or more agents that autonomously trigger and settle payments — usage-based services, access control, dynamic pricing |
| **T-Pay** | Agentic / x402 Payments | ⊢ AI assistant that makes payments on a user's behalf with built-in rules — spending limits, approval checkpoints, KYC/AML |
| **T-Commerce** | Web Data Commerce / Financial Ops | ⊢ Digital product/service with x402-based revenue model — token-gated access, real-time rev-splits, instant payouts; or business tooling for real-time payments + audit-ready records |
| **T-Web** | General Web Data | ⊢ Agents, data pipelines, search workflows, scraping systems, production-ready AI infrastructure |

⊨ LedgerLens claims **T-Agent (primary)** + **T-Commerce (secondary)** + **T-Web (tertiary depth)**.

⊥ Unknown: exact judging weights, judges' names, per-track prizes. **Assumed criteria** (carried from prior Bright Data hackathon page): Application of Technology · Presentation · Business Value · Originality. Re-verify at T+0 kickoff stream.

### 2.2 x402 context

```
x402 ≜ [
  origin:          Coinbase (revives HTTP 402 "Payment Required"),
  rail:            USDC micropayments on Base (Sepolia for testnet),
  official_SDKs:   TypeScript · Go · Python (tier-1 reference implementations),
  facilitator:     Coinbase CDP facilitator — production, free tier 1,000 tx/month,
  foundation:      x402 Foundation = Coinbase + Cloudflare,
  adoption_2026:   Stripe (Feb), AWS Bedrock AgentCore Payments (May),
                   Google Agentic Payments + x402, Solana 35M+ tx
]
```

---

## 3. The imagined goal — Alchy step 1–3

```
Imagined_Goal ≜
  "Be the team that makes the trust layer of the agentic commerce stack
   visible and inevitable. Win the $5K cash, secure the Startup Program admission,
   walk out of the SF Web Data Loft with a positioning that survives the demo:
   'GEM² is the epistemic gate every agentic payment will eventually pass through.'"

Desired_Future_State ≜ [
  prize:            grand prize OR top-3 honors,
  startup_program:  admitted with credits committed,
  positioning:      GEM² publicly tied to "trust layer for agentic commerce",
  pipeline:         ≥3 enterprise conversations from onsite presence,
  artifact:         a real runnable product that survives as v0 of a startup
]

Inferred_Real_Intention ≜
  "David is not shipping a hackathon project. He is seeding a product line
   where GEM² is the audit-and-trust layer for the emerging agentic-commerce +
   live-web-data substrate. The hackathon is the launchpad, the proof,
   and the network entry."
```

⊨ Optimize for **two horizons simultaneously**:
- **H1 (8 days):** maximize judging score across the four criteria.
- **H2 (90 days):** leave behind an asset that can be raised against, sold into, or evolved.

The Go-first stack and visible Trust Gate are calibrated to hit both.

---

## 4. Product architecture — `LedgerLens`

### 4.1 Canonical positioning sentences

**English headline (for README, deck, video):**
> LedgerLens is an x402-native Agent-to-Agent Payments product with a Bright Data-powered web evidence layer and a GEM² Trust Gate before settlement.

**Korean headline:**
> LedgerLens는 AI 에이전트 간 웹 데이터 거래에서 Bright Data로 근거를 수집하고, GEM² Trust Gate가 검증한 경우에만 x402 결제를 허용하는 신뢰 기반 결제 게이트다.

**The judge sentence (final slide / video close):**
> LedgerLens makes autonomous agent payments safe by forcing every web-data purchase through live evidence verification and a GEM² Trust Gate before x402 settlement.

**Slogan everywhere:**
> **No grounded claim, no payment.**

### 4.2 The contract

```
F_LedgerLens : A → B | P

A ≜ [
  buyer_request:     BuyerRequest,
  seller_offers:     Seq(SellerOffer),
  policy:            PaymentPolicy [spend_cap, public_only, claim_grounded_required],
  bright_data_creds: Path,
  x402_facilitator:  URL
]

B ≜ [
  matched_offer:     SellerOffer,
  evidence_set:      Seq(EvidenceReceipt),          (* L0 *)
  claim_assessments: Seq(ClaimAssessment),          (* L2 — canonical 4 tags *)
  decision_packet:   DecisionPacket,                (* L3 *)
  settlement:        X402Settlement,                (* approved path only *)
  audit_bundle_path: Path                           (* exportable *)
]

P ≜ buyer_request ≠ ⊥
    ∧ |seller_offers| ≥ 1
    ∧ gem2-tpmn-checker.fly.dev reachable (audit gate)
    ∧ Bright Data credentials valid
    ∧ policy.spend_cap > 0
    ∧ settlement_mode = "simulation"   (* hackathon lock — no chain, no keys *)
```

### 4.3 Why this wins all four assumed criteria

| Criterion | How LedgerLens scores |
|---|---|
| **Application of Tech** | 4 Bright Data products in distinct roles + production `gem2-tpmn-checker` audit gate (L1 P-check + L2 O-check) + visible x402 protocol simulation + visible GEM² L0–L3 architectural stack. Most submissions will use 1 Bright Data product; we use 4. Most teams will skip the verification layer entirely; we leverage a deployed production audit gate. |
| **Presentation** | Single narrative: "agents shouldn't pay for claims they can't verify." Demo shows the gate firing in both directions. |
| **Business Value** | Real enterprise pain — CFOs cannot approve agentic spend because models hallucinate. Solve trust, unlock spend. Direct buyer (CFO / Head of AI). |
| **Originality** | Epistemic gate before payment is structurally unique. The visible SPT guards (S→T, L→G) are quotable demo moments. |

---

## 5. The GEM² Trust Gate — architecture and rules

LedgerLens uses **two complementary GEM² framings simultaneously**:

1. **Architectural stack** (high-level) — four layers describing the audit pipeline as a whole.
2. **Per-contract gate pair** (production canonical) — L1 P-check + L2 O-check, as deployed at `gem2-tpmn-checker.fly.dev`.

This dual framing is honest because:
- ⊢ The deep-research framing (Evidence → Memory → Verification → Release) is the architectural stack.
- ⊢ The production canonical (`/api/audit-gate/p-check`, `/api/audit-gate/o-check`) is the gate-pair INSIDE the Verification layer of the stack.
- ⊨ LedgerLens integrates the **deployed audit-gate API** — we do not rebuild it. This eliminates the L2 prompt-engineering risk and brings production-grade RAG-augmented epistemic verification on Day 1.

### 5.1 Architectural stack (high-level)

| Layer | Purpose | Rule | Artifacts |
|---|---|---|---|
| **Evidence Layer** | Forensic acquisition of source material via Bright Data | **If there is no source receipt, no claim may proceed.** | `fetch_receipts.jsonl`, raw URL, timestamp, Bright Data product used, response hash, raw HTML/markdown/screenshot, source metadata |
| **Memory Layer** | Entity, offer, seller, buyer, source memory | **If entity identity or source context is ambiguous, downgrade the claim.** | seller profile, buyer request log, offer memory, source reputation, prior tx history, SQLite entity graph (mirrors TechEx `layer_audit_log` schema) |
| **Verification Layer** | Claim scoring + overclaim detection — **implemented via the deployed audit-gate API** | **Unsupported or speculative claims cannot trigger payment approval.** | L1 P-check call + L2 O-check call to `gem2-tpmn-checker.fly.dev` with evidence-augmented requests; reasons parsed into `claim_assessments[]` with canonical EEF tag, SPT violations, EVIDENCE-N correlation, confidence |
| **Release Layer** | Governance boundary before x402 settlement — local Go composite policy | **x402 settlement is allowed only after composite L1∧L2 + policy release.** | `decision_packet.json` (includes both upstream `result_id`s), policy evaluation log, payment authorization record, audit bundle |

### 5.2 Per-contract gate pair (production canonical)

The Verification Layer in §5.1 is realized by calling two production endpoints, each evaluating one half of a TPMN contract `F: A → B | P`:

```
For every (BuyerRequest, SellerOffer) candidate pair:

  ① L1 P-check (BEFORE matching)
     POST gem2-tpmn-checker.fly.dev/api/audit-gate/p-check
     {
       i: <seller offer text>,
       a: "SellerOffer{offerId, claim, priceUSDC, sourceUrl}",
       p: [
         "claim must be supported by the provided evidence",
         "no SPT violations (S→T, L→G, Δe→∫de)",
         "source URL must be public and reachable",
         "price ≤ buyer.maxSpendUSDC"
       ],
       t: 70,
       evidence: [<Bright Data EvidenceReceipt content chunks>],
       gem2_api_key: ⊥,
       gemini_api_key: ⊥
     }
     → {verdict: ALLOW|DENY, score, reasons: [EVIDENCE-N, EEF-⊢/⊨/⊬/⊥, SPT, RULE-N, DIM]}

  ② Local F: match buyer ↔ offer; compose draft DecisionPacket

  ③ L2 O-check (AFTER matching)
     POST gem2-tpmn-checker.fly.dev/api/audit-gate/o-check
     {
       o: <draft DecisionPacket>,
       b: "DecisionPacket{verdict, claimAssessments, paymentAllowed, ...}",
       p: [
         "decisionPacket.verdict is consistent with claimAssessments",
         "paymentAllowed = true iff verdict = APPROVED",
         "totalCost ≤ buyer.maxSpendUSDC (conservation of policy)"
       ],
       t: 75,
       evidence: [<same evidence corpus>]
     }
     → {verdict: SUCCESS|FAILURE, score, reasons: [...]}

  ④ L3 Release (local Go)
     APPROVED  iff  L1.ALLOW ∧ L1.score≥70 ∧ L2.SUCCESS ∧ L2.score≥75 ∧ within_policy
     BLOCKED   iff  L1.DENY  ∨ L2.FAILURE ∨ critical_⊬-without-basis ∨ critical_⊥
     ESCALATE  iff  intermediate-band scores with no hard fail
     → x402 settlement iff APPROVED
```

⊨ **Trust principle (borrowed from TechEx):** the judge is NEVER the same LLM as the worker. The audit gate uses Gemini-3-flash-preview (or Claude/OpenAI per request); any LLM we use locally for buyer/seller agent dialogue is different. That separation is why the verdict is trustworthy.

### 5.3 Optional Lobster Trap defense (regex-only, T+2 stretch)

⊢ TechEx ships a pure-Go regex DPI engine (`console/lobstertrap.go`, MIT) that scans content for prompt-injection patterns BEFORE any LLM call. LedgerLens can borrow this to scan scraped seller pages, adding cheap pre-flight defense without adding LLM cost.

| Position | What | LedgerLens usage |
|---|---|---|
| **L0 ingress** | Regex scan of fetched HTML/markdown for injection attempts (data exfil, role override, special-instruction blocks) | Run on every Bright Data Unlocker/Browser fetch before content reaches the audit gate's `evidence: []string` |
| **L3 egress** | Regex scan of rendered decision packet text | Run on the audit bundle before public export |

⊨ Decision: ship Lobster Trap **regex-only** (no LLM canonicalize) as a T+2 stretch goal. Cuttable. If shipped, demo gets a "we already thought of that" answer for prompt injection.

### 5.4 Claim taxonomy — canonical EEF (4 tags), with UI enrichment

```
Canonical_EEF (API + audit bundle) ≜ {
  ⊢ grounded     — confirmed by direct evidence,
  ⊨ inferred     — derived from grounded claims with visible chain,
  ⊬ extrapolated — beyond evidence; basis must be stated (or absent → "speculative" UI label),
  ⊥ unknown      — knowledge gap; stops inference chain
}

UI_Label (display only) ≜ {
  "grounded"    ↔ ⊢,
  "inferred"    ↔ ⊨,
  "extrapolated" ↔ ⊬ with stated basis,
  "speculative"  ↔ ⊬ with NO stated basis (UI emphasis only — NOT a 5th tag),
  "unknown"     ↔ ⊥
}
```

⊢ Canonical 4 tags stay in `audit_bundle.json` and the public API. "Speculative" is a UI affordance that makes the worst subclass of ⊬ visible to non-technical viewers.

### 5.5 SPT guardrails (three classes demo-able live)

| SPT class | LedgerLens demo example |
|---|---|
| **S→T** State-as-Trait | Seller: "Our API has 99.9% uptime" (one week of data). Gate marks ⊬, demands a longer evidence window. |
| **L→G** Local-as-Global | Seller: "Universally accurate pricing data." Bright Data fetch shows US-only coverage. Gate narrows claim, marks ⊬. |
| **Δe→∫de** Increment-as-Mass | Seller: "+12% MoM growth proves the trend." Two data points. Gate refuses "trend" inference. |

### 5.6 Thin compliance surface (the word stays honest)

⊢ Five concrete compliance features that ship in the MVP, so "Trust Gate" survives any judge who asks about compliance:

| Surface | What | File / artifact |
|---|---|---|
| Spending cap | Per-request `max_spend_usdc` enforced at L3 | `policy.spend_cap`, blocked transactions logged |
| Public-only collection | Bright Data calls scoped to public sources; no login attempts | `LEGAL.md` + `internal/brightdata/policy.go` |
| Immutable audit log | Append-only `audit_bundle.json` with hash chain | `artifacts/audit_bundles/` |
| AUP-aware source policy | Bright Data AUP boundaries encoded as Go predicates | `internal/brightdata/aup.go` |
| Decision packet | Every payment (approved or blocked) produces a packet | `decision_packet.json` |

---

## 6. Stack lock — Go-first

### 6.1 Ground for the lock (Kritik-verified 2026-05-23)

```
⊢ Coinbase x402: Go is a tier-1 official reference implementation
    (coinbase/x402/go) alongside TypeScript and Python,
⊢ Coinbase x402 Go exports: X402Client, X402ResourceServer, X402Facilitator,
⊢ USDC + Base Sepolia constants pre-configured (x402.BaseSepolia),
⊢ mark3labs/mcp-go: 1,880 importers, dominant Go MCP SDK,
⊢ mark3labs/mcp-go-x402 EXISTS: x402 payment transport for MCP-Go
    = the canonical pattern for "agent pays to use an MCP tool"
    = precisely LedgerLens's core loop,
⊨ Stack friction risk that motivated initial caution: not present.
```

### 6.2 Principle

> **Go owns truth. TypeScript owns UI and optional SDK convenience.**

The final `DecisionPacket`, `PaymentAllowed`, `AuditBundle`, and `GateVerdict` are **always** produced by Go. TS shims are bridges, not authorities.

### 6.3 Repository structure (collapsed from "4 services" → "1 binary")

```text
apps/
  web/                          # Next.js + Tailwind — demo UI

cmd/
  ledgerlens/                   # ONE Go binary, all internal packages below

internal/
  trustgate/                    # L0/L1/L2/L3 audit logic
    l0_evidence/                # fetch receipts, hashing, raw storage
    l1_memory/                  # entity graph, retrieval index, sqlite store
    l2_verify/                  # claim scoring, SPT checks, EEF tagging
    l3_release/                 # gate policy, decision packet, ValidHandoff
  paymentgate/                  # x402 policy + settlement state machine
  brightdata/                   # SERP, Unlocker, Browser, MCP-client wrappers
    aup.go                      # AUP-aware source policy
    policy.go                   # public-only enforcement
  schemas/                      # canonical Go structs — single source of truth
  agents/                       # role contracts (buyer, seller, verifier, gate)

adapters/
  mcp-bridge/                   # (RESERVED, likely unused) tiny TS process
  x402-signer/                  # (RESERVED, likely unused) tiny TS helper

packages/
  contracts-ts/                 # TS types generated from Go structs

storage/
  migrations/                   # sqlite schema migrations
  seeds/                        # demo seller fixtures

artifacts/
  audit_bundles/                # exported bundles
  fetch_receipts/               # raw L0 receipts
  demo_cases/                   # seeded blocked + approved scenarios

docs/
  Bright-Data-winning-strategy.md   # this doc (canonical)
  BRIGHTDATA_INTEGRATION.md         # per-product role + code pointer
  GEM2_AUDIT_MODEL.md               # L0-L3 spec + EEF mapping
  X402_FLOW.md                      # payment state machine
  DEMO_SCRIPT.md                    # 5-min demo storyboard
  JUDGE_GUIDE.md                    # "look here first" for judges
  LEGAL.md                          # public-only scope, AUP, Sepolia
  THIRD_PARTY_NOTICES.md            # MIT + CC-BY-4.0 attribution

.gem-squared/
  alarm.md, work-plan/, verify-work-logs/, truth-logs/, archive/
```

### 6.4 Concrete dependencies

| Layer | Library / endpoint | Why |
|---|---|---|
| HTTP server | `github.com/gin-gonic/gin` or `net/http` + chi | TechEx also uses gin |
| **x402 (hackathon)** | **none** — local Go state machine modeling HTTP 402 lifecycle | **v2.2 lock — SIMULATION mode for public demo safety.** No SDK, no chain, no private keys. |
| x402 (post-hackathon, NOT in this repo) | `github.com/coinbase/x402/go` (official Go SDK) — swap behind `Settler` interface | Tier-1 reference implementation available when David is ready to ship real settlement |
| MCP client | `github.com/mark3labs/mcp-go` | Production Go MCP SDK (1,880+ importers) |
| MCP+x402 bridge (post-hackathon) | `github.com/mark3labs/mcp-go-x402` | Direct pattern fit for future MCP-tool-pay |
| Bright Data HTTP | stdlib `net/http` + thin wrapper | SERP/Unlocker/Browser are all REST |
| **Audit gate** | `gem2-tpmn-checker.fly.dev` (production HTTP API, v1.1) | **NO SDK needed — stdlib `net/http`. GEM².AI operates the platform; `GEM2_API_KEY` already provisioned.** Eliminates L2 prompt-engineering risk. |
| Audit gate Go client | port from `TechEx/console/audit_gate_client.go` (MIT) | Working reference; copy-paste viable |
| Storage | SQLite via `modernc.org/sqlite` (pure-Go) | Zero-dep demo; matches TechEx stack exactly |
| Layer audit log schema | port from `TechEx/console/audit_log.go` `layer_audit_log` table (MIT) | Regulator-replay-ready local mirror |
| LLM provider key | `GEMINI_API_KEY` OR `ANTHROPIC_API_KEY` (per-request, passed to audit gate) | Gate is LLM-agnostic; we don't host long-lived LLM keys |
| Lobster Trap (optional T+2 stretch) | port from `TechEx/console/lobstertrap.go` (pure-Go regex, MIT) | Prompt-injection defense without LLM cost |
| Hashing / signing | stdlib `crypto/sha256` + `go-ethereum/crypto` if needed | Standard |
| Web | Next.js 15 + React + Tailwind | Frontend only |
| Schema sharing | `tygo` (Go → TS type generator) | Single source of truth |

⊢ TS shim slots stay in the tree but only get filled if a friction surface forces it. Expectation: zero TS code outside `apps/web/`.

### 6.5 Canonical schemas (Go is the source; TS is generated)

```go
// internal/schemas/types.go

type SellerOffer struct {
    OfferID    string  `json:"offerId"`
    SellerID   string  `json:"sellerId"`
    Claim      string  `json:"claim"`
    PriceUSDC  float64 `json:"priceUSDC"`
    SourceURL  string  `json:"sourceUrl,omitempty"`
    CreatedAt  string  `json:"createdAt"`
}

type BuyerRequest struct {
    RequestID     string        `json:"requestId"`
    BuyerID       string        `json:"buyerId"`
    Query         string        `json:"query"`
    MaxSpendUSDC  float64       `json:"maxSpendUSDC"`
    Policy        PaymentPolicy `json:"policy"`
}

type EvidenceReceipt struct {
    ReceiptID         string `json:"receiptId"`
    URL               string `json:"url"`
    BrightDataProduct string `json:"brightDataProduct"` // SERP|UNLOCKER|BROWSER|SCRAPER|MCP
    FetchedAt         string `json:"fetchedAt"`
    ContentHash       string `json:"contentHash"`
    RawRef            string `json:"rawRef"`
}

type ClaimStatus string
const (
    ClaimGrounded     ClaimStatus = "grounded"     // ⊢
    ClaimInferred     ClaimStatus = "inferred"     // ⊨
    ClaimExtrapolated ClaimStatus = "extrapolated" // ⊬ (UI may render as "speculative" when basis is absent)
    ClaimUnknown      ClaimStatus = "unknown"      // ⊥
)

type ClaimAssessment struct {
    ClaimID        string      `json:"claimId"`
    Claim          string      `json:"claim"`
    Status         ClaimStatus `json:"status"`
    Basis          string      `json:"basis,omitempty"`     // required if Extrapolated
    EvidenceRefs   []string    `json:"evidenceRefs"`
    SPTViolations  []string    `json:"sptViolations"`       // "S->T" | "L->G" | "Δe->∫de"
    Confidence     float64     `json:"confidence"`
}

type GateVerdict string
const (
    GateApprovedByTrustGate GateVerdict = "APPROVED_BY_TRUST_GATE"
    GateBlockedByTrustGate  GateVerdict = "BLOCKED_BY_TRUST_GATE"
    GateEscalatedToHuman    GateVerdict = "ESCALATED_TO_HUMAN"
)

type DecisionPacket struct {
    DecisionID        string            `json:"decisionId"`
    RequestID         string            `json:"requestId"`
    OfferID           string            `json:"offerId"`
    Verdict           GateVerdict       `json:"verdict"`
    Reason            string            `json:"reason"`
    ClaimAssessments  []ClaimAssessment `json:"claimAssessments"`
    PaymentAllowed    bool              `json:"paymentAllowed"`
    AuditBundleRef    string            `json:"auditBundleRef"`
    L1ResultID        string            `json:"l1ResultId,omitempty"` // gem2-tpmn-checker upstream result_id
    L2ResultID        string            `json:"l2ResultId,omitempty"`
}

// x402 SIMULATION mode (v2.2) — no chain, no private keys, no real funds
type PaymentState string
const (
    PaymentRequired         PaymentState = "PAYMENT_REQUIRED"
    PendingVerification     PaymentState = "PENDING_VERIFICATION"
    ApprovedByTrustGate     PaymentState = "APPROVED_BY_TRUST_GATE"
    BlockedByTrustGate      PaymentState = "BLOCKED_BY_TRUST_GATE"
    EscalatedToHuman        PaymentState = "ESCALATED_TO_HUMAN"
    SimulatedSettled        PaymentState = "SIMULATED_SETTLED"
)

type SettlementMode string
const (
    ModeSimulation SettlementMode = "simulation"        // hackathon default
    ModeMainnet    SettlementMode = "mainnet"           // post-hackathon — not used here
    ModeTestnet    SettlementMode = "testnet"           // post-hackathon — not used here
)

type SimulatedSettlement struct {
    SettlementID     string         `json:"settlementId,omitempty"` // null when blocked
    DecisionID       string         `json:"decisionId"`
    Mode             SettlementMode `json:"mode"`                    // always "simulation" in hackathon repo
    Network          string         `json:"network"`                 // always "demo-local"
    Asset            string         `json:"asset"`                   // always "USDC-demo"
    Status           PaymentState   `json:"status"`
    AmountUSDC       float64        `json:"amountUSDC"`              // 0 when blocked
    Reason           string         `json:"reason"`
    RealTransaction  bool           `json:"realTransaction"`         // always false
    PrivateKeysUsed  bool           `json:"privateKeysUsed"`         // always false
    RealFundsUsed    bool           `json:"realFundsUsed"`           // always false
    Ts               string         `json:"ts"`                      // ISO8601
}

// Settler is the interface that a real Coinbase x402 implementation would satisfy
// post-hackathon. The hackathon repo ships ONLY SimulatedSettler.
type Settler interface {
    Settle(ctx context.Context, decision DecisionPacket) (SimulatedSettlement, error)
}
```

---

## 7. Bright Data integration plan

| Product | Role | Code path | Demo moment |
|---|---|---|---|
| **SERP API** | Discovery | `internal/brightdata/serp.go` | Finds candidate evidence pages for a seller's claim |
| **Web Unlocker** | Static retrieval | `internal/brightdata/unlocker.go` | Fetches seller's trust / pricing / docs page |
| **Browser API** | Interactive / JS-heavy | `internal/brightdata/browser.go` | Verifies a JS-heavy dashboard claim live |
| **MCP Server** | Live buyer-agent investigation | `internal/brightdata/mcp_client.go` (via `mark3labs/mcp-go`) | Buyer agent invokes Bright Data tools mid-conversation |
| **Web Scraper API** (optional) | Structured extraction | `internal/brightdata/scraper.go` | LinkedIn / Crunchbase / GitHub seller profile sanity check |

⊢ Mandatory `docs/BRIGHTDATA_INTEGRATION.md` documents each product with role, code pointer, and a sample fetch receipt. This is unforced sponsor depth — never skip.

---

## 8. x402 protocol simulation flow (v2.2 — no real chain)

⊢ Hackathon ships in **simulation mode**. The state machine and HTTP-402 shape are real; the settlement layer is a deterministic local stub. This protects David's personal Coinbase account from public-demo exposure and concentrates the demo on what actually matters: the L3 Trust Gate verdict.

```
HTTP 402-shaped lifecycle (simulated):

PAYMENT_REQUIRED
    │  buyer_agent receives an HTTP 402-shaped challenge with payment requirements
    ▼
PENDING_VERIFICATION
    │  Evidence Layer: Bright Data fetches receipts
    │  Memory Layer:   entity/source memory updated
    │  L1 P-check:     POST gem2-tpmn-checker.fly.dev/api/audit-gate/p-check
    │  F (local):      compose draft DecisionPacket
    │  L2 O-check:     POST gem2-tpmn-checker.fly.dev/api/audit-gate/o-check
    │  L3 Release:     composite verdict from L1 ∧ L2 ∧ policy
    │
    ├─ APPROVED_BY_TRUST_GATE
    │       │  SimulatedSettler emits deterministic receipt:
    │       │    {settlement_id: sim_x402_<id>, mode: "simulation",
    │       │     network: "demo-local", asset: "USDC-demo",
    │       │     real_transaction: false, ...}
    │       ▼
    │  SIMULATED_SETTLED  → audit bundle exported with simulated settlement receipt
    │
    ├─ BLOCKED_BY_TRUST_GATE
    │       │  No payment authorization issued.
    │       │    {settlement_id: null, status: "BLOCKED_BY_TRUST_GATE",
    │       │     amount_usdc: 0, reason: "<L1/L2 reason chain>"}
    │       ▼
    │  audit bundle exported with reason
    │
    └─ ESCALATED_TO_HUMAN
            │  intermediate-band scores; held for operator decision
            ▼
       audit bundle exported with held state
```

### 8.1 The transparency badge

⊢ Every settlement record carries `mode: "simulation"` and `real_transaction: false`. The UI shows a small **`SIMULATION MODE`** pill near every settlement event. This is a posture, not a weakness: it explicitly states *"we kept your private key out of our demo."*

### 8.2 The pitch line

> **"For demo safety, we do not connect a personal Coinbase account or expose private keys. Instead, we simulate the x402 payment lifecycle and focus on the core innovation: the GEM² Trust Gate decides whether an autonomous agent is allowed to pay."**

Shorter:

> **"We simulate settlement, but the trust gate is real."**

### 8.3 Real x402 — post-hackathon path

⊨ The `Settler` interface (see §6.5) is the swap point. Post-hackathon, replace `SimulatedSettler` with a `CoinbaseX402Settler` that wraps the official Coinbase Go SDK + CDP facilitator + Base Sepolia (later: mainnet). The Trust Gate code does not change. This is the cleanest possible boundary between "demo-safe" and "production-real."

⊨ Payment is the **consequence** of verification, never the centerpiece. The Trust Gate is the centerpiece.

---

## 9. Eight-day execution arc

| Day | Calendar | Goal | Concrete deliverables |
|---|---|---|---|
| **T-2** | Sat 05-23 (today) | Strategy v2 locked; repo scaffold; stack locked | This doc; repo + MIT license; `.gem-squared/` initialized; `cmd/ledgerlens` skeleton |
| **T-1** | Sun 05-24 | Go schemas + module init | `internal/schemas/types.go`; `go.mod`; `tygo` for TS generation; Next.js scaffold |
| **T+0** | Mon 05-25 | **Kickoff** — claim $250 credits; verify judging weights; lock orchestration choice | Discord join; kickoff stream; product name confirmed publicly |
| **T+1** | Tue 05-26 | Bright Data ingestion working | `internal/brightdata/{serp,unlocker,browser,mcp_client}.go` returning real receipts; 5 seeded seller fixtures |
| **T+2** | Wed 05-27 | GEM² L0/L1/L2 producing ClaimAssessments | `internal/trustgate/{l0,l1,l2}/*.go`; first ⊢/⊨/⊬/⊥ output from real evidence |
| **T+3** | Thu 05-28 | x402 simulation + L3 gate end-to-end | `internal/paymentgate/*.go`; `SimulatedSettler` wired; deterministic `sim_x402_<id>` receipts visible; decision packets produced with L1/L2 result_ids |
| **T+4** | Fri 05-29 | UI + audit panel | Next.js dashboard with buyer view · seller cards · evidence panel · claim assessments · settlement transcript |
| **T+5** | Sat 05-30 | **HYBRID DAY (onsite SF)** · scenario polish · video shoot | 3 seeded demo scenarios stable; smoke test live; ≤5-min video recorded; deck v1; arrive at Web Data Loft |
| **T+6** | Sun 05-31 (AM) | Submission pack | Public repo flip; live URL stable; deck final; `BRIGHTDATA_INTEGRATION.md` + `AUDIT_BUNDLE_SAMPLE.json` + `LEGAL.md` + `THIRD_PARTY_NOTICES.md` |
| **T+6** | Sun 05-31 (PM) | **DEMOS & AWARDS** | Present onsite; network at Web Data Loft; ≥3 enterprise conversations |

⊢ **Hard rule — feature freeze T+5 EOD.** T+6 AM is submission polish only.

⊨ **Scope-cut order if 8 days proves tight** (cut in this sequence):
1. Web Scraper API → drop to 3 Bright Data products (still strong)
2. SPT class #3 (Δe→∫de) — keep S→T and L→G as the demo classes
3. T-Commerce track claim — narrow to T-Agent + T-Web
4. Web UI fanciness (charts, animation) — keep the audit panel only
5. **NEVER cut:** L2 audit panel · x402 protocol simulation · blocked-payment demo case · ≤5-min video · MIT license · SIMULATION MODE badge in UI

---

## 10. Demo narrative — five minutes that win the room

### 10.1 Slide spine (8 slides matching directive)

| # | Slide | Content (≤15 words) |
|---|---|---|
| 1 | **Hook** | "Autonomous agents are about to spend trillions. Today they hallucinate. Who guards the wallet?" |
| 2 | **The problem** | Agents pay for claims they cannot verify. CFOs cannot approve agentic spend. |
| 3 | **The solution** | LedgerLens — the GEM² Trust Gate between an agent and its wallet. |
| 4 | **Product flow** | seller offer → buyer request → web verification → GEM² audit → x402-shaped simulated settlement |
| 5 | **Bright Data architecture** | SERP + Unlocker + Browser + MCP routing diagram |
| 6 | **GEM² L0–L3 audit architecture** | layer-by-layer with the four rules and the EEF tags |
| 7 | **Live demo** | [switch to live app — both cases below] |
| 8 | **Business value + why we win** | trust layer for agentic commerce · not another payment demo · auditable governance |

### 10.2 Live demo — two cases

#### Case A — Blocked Payment (the wow moment)

```
Seller Agent claims:
  "This dataset contains verified real-time pricing from the official vendor site."

LedgerLens flow:
  ├─ Evidence Layer: Bright Data SERP discovers seller's evidence page
  ├─                  Bright Data Unlocker fetches the page
  ├─                  receipt with URL + hash + timestamp + product=UNLOCKER
  ├─ Memory Layer:   seller has no prior verified transactions
  ├─ L1 P-check (gem2-tpmn-checker):
  │       evidence: stale 6-month-old mirror
  │       claim:    "real-time pricing from official vendor"
  │       → verdict = DENY, score = 38
  │       → reasons: [EVIDENCE-1] outdated snapshot — supports ⊬ extrapolation,
  │                  [SPT-S→T] outdated snapshot presented as current state,
  │                  [EEF-⊬] basis: "data freshness cannot be certified from stale evidence",
  │                  [RULE-1] FAIL: source is 6 months stale
  ├─ L3 Release:    BLOCKED_BY_TRUST_GATE (L1 DENY → hard block; L2 not invoked)
  └─ Simulated Settler:
        {settlement_id: null,
         status: "BLOCKED_BY_TRUST_GATE",
         mode: "simulation",
         amount_usdc: 0,
         real_transaction: false,
         reason: "Seller claim not grounded; evidence 6 months stale"}

UI banner (red):
  "Payment blocked: seller claim not grounded.
   Evidence is 6 months stale; cannot certify 'real-time'."

Side panel shows the L1 reason chain with the [SPT-S→T] badge.

Audit bundle exports with full reason chain + null settlement_id.
```

⊨ This is the case judges will quote in the awards announcement. Practice it.

#### Case B — Approved Payment (simulated settlement)

```
Seller Agent claims:
  "Live NYSE+NASDAQ price feed, 1-second freshness, $0.001 per query."

LedgerLens flow:
  ├─ Evidence Layer: Bright Data SERP discovers seller status page
  ├─                  Bright Data Browser API verifies live JS-rendered dashboard
  ├─                  receipt with URL + hash + timestamp + product=BROWSER
  ├─ Memory Layer:   seller has 12 prior simulated-approved transactions, all under cap
  ├─ L1 P-check (gem2-tpmn-checker):
  │       evidence: live dashboard render + status page
  │       claim broken into 3 subclaims:
  │         "NYSE+NASDAQ coverage" → [EEF-⊢] grounded (status page lists both)
  │         "1-second freshness"   → [EEF-⊢] grounded (render timestamp aligns)
  │         "$0.001 per query"     → [EEF-⊢] grounded (pricing page)
  │       → verdict = ALLOW, score = 94
  ├─ F (local):     compose draft DecisionPacket
  ├─ L2 O-check (gem2-tpmn-checker):
  │       postconditions:
  │         "decisionPacket.verdict consistent with claimAssessments" → PASS
  │         "paymentAllowed = true iff verdict = APPROVED"            → PASS
  │         "totalCost ≤ buyer.maxSpendUSDC"                          → PASS
  │       → verdict = SUCCESS, score = 96
  ├─ L3 Release:    APPROVED_BY_TRUST_GATE (composite green)
  └─ Simulated Settler:
        {settlement_id: "sim_x402_a3f17c92",
         decision_id:  "<uuid>",
         mode:         "simulation",
         network:      "demo-local",
         asset:        "USDC-demo",
         status:       "SIMULATED_SETTLED",
         amount_usdc:  0.001,
         real_transaction: false,
         private_keys_used: false,
         real_funds_used:   false,
         reason:       "L3 Trust Gate approved grounded claim",
         ts:           "2026-05-31T17:42:11Z"}

UI banner (green):
  "Payment approved: all claims grounded.
   Simulated settlement issued — sim_x402_a3f17c92."

Side panel shows L1 + L2 result_ids (regulator-replayable) + the SIMULATION MODE pill.

Audit bundle exports with simulated settlement receipt + L1/L2 reason chains.
```

⊨ Case B is intentionally NOT about a flashy on-chain explorer link. It is about a complete, auditable, gate-driven payment lifecycle that judges can replay deterministically.

---

## 11. Submission deliverables — anti-failure checklist

| Artifact | Mandatory | Quality bar | Owner |
|---|---|---|---|
| Public MIT-licensed GitHub repo | ✓ | README with the canonical headline (§4.1) + judge-guide link + the **`## x402 Demo Mode`** block below (§11.1) | David |
| Live application URL | ✓ | Deployed; seeded scenarios available even if live web flaps | Gineer |
| Video ≤5 min | ✓ | Demo first, slides second; both cases shown; captioned | David |
| Slide deck | ✓ | 8-slide spine in §10.1 | Alchy/David |
| Title, summary, long description, tags | ✓ | Tags: `x402`, `agentic-payments`, `bright-data`, `mcp`, `trust-gate`, `gem2`, `web-data-commerce` | David |
| Cover image | ✓ | The blocked-payment UI banner — not a logo | Designer / generated |
| `docs/BRIGHTDATA_INTEGRATION.md` | recommended | Per-product role · code pointer · sample receipt | Gineer |
| `docs/GEM2_AUDIT_MODEL.md` | recommended | L0–L3 spec + canonical EEF + UI label mapping | Alchy |
| `docs/X402_FLOW.md` | recommended | Simulation state machine + `Settler` interface + transparency rationale + post-hackathon real-x402 swap path | Gineer |
| `docs/DEMO_SCRIPT.md` | recommended | Both demo cases word-for-word | David |
| `docs/JUDGE_GUIDE.md` | recommended | 5-bullet "look here first" | Alchy |
| `LEGAL.md` | recommended | Public-only scope · Bright Data AUP · Sepolia-only · spending caps | Alchy |
| `THIRD_PARTY_NOTICES.md` | required if reusing TPMN-PSL text | MIT for own code · CC-BY-4.0 attribution for any TPMN spec excerpts | David |
| `artifacts/AUDIT_BUNDLE_SAMPLE.json` | recommended | A real exported bundle from Case A and Case B | Gineer |
| `eval/results.md` + gold set | recommended | 10 seeded scenarios; precision + grounded-claim ratio numbers | Gineer |

⊬ Disqualification traps:
- MIT license sloppiness (mixing CC-BY-4.0 text into MIT-claimed code without attribution)
- Forgetting to register on lablab + form the team object (even solo entries register as a team)
- Submitting before `BRIGHTDATA_INTEGRATION.md` is present (the easiest App-of-Tech evidence)

### 11.1 README — `## x402 Demo Mode` block (canonical text)

Copy verbatim into the public README:

```markdown
## x402 Demo Mode

LedgerLens demonstrates an x402-compatible payment flow in **simulation mode**.

For public demo safety, **no real private keys, Coinbase accounts, or
mainnet/testnet funds are used**. The system models the x402 payment lifecycle:

1. Payment required (HTTP 402-shaped challenge)
2. Bright Data evidence collection
3. GEM² L1 P-check + L2 O-check via the deployed `gem2-tpmn-checker.fly.dev` audit-gate API
4. L3 Trust Gate decision (composite verdict)
5. Payment authorization or block
6. Simulated settlement receipt (tagged `mode: "simulation"`, `real_transaction: false`)

The purpose of the demo is to prove the **governance layer before settlement**:

> **No grounded claim, no payment.**

A real x402 settlement implementation (Coinbase Go SDK on Base Sepolia → mainnet)
is a post-hackathon swap behind the `Settler` interface. The Trust Gate code
does not change.
```

---

## 12. Risks + mitigations

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Bright Data API credit exhaustion mid-build | medium | high | Aggressive Evidence Layer cache; MCP free tier (5K/mo) for live-agent paths; request more via Discord on T+2 |
| ~~Base Sepolia faucet / Coinbase CDP risk~~ | — | — | **ELIMINATED in v2.2** — settlement is simulation; no chain, no faucet, no CDP signup. Real x402 deferred to post-hackathon. |
| **gem2-tpmn-checker.fly.dev unreachable mid-demo** (v2.3 new) | low-medium | high | Cache audit-gate responses for every seeded scenario into `artifacts/demo_cases/*.json`; UI exposes a **LIVE / REPLAY** toggle (already in §10 spec). On API failure, auto-fall-back to replay with a visible "REPLAY MODE" pill alongside SIMULATION MODE. Cached responses include the full `[EEF-…]`/`[SPT-…]`/`[EVIDENCE-N]` reason chain so the demo narrative is identical. Staging fallback URL also configured. |
| Live Bright Data source flaps mid-presentation | medium | medium | Same seeded scenarios in `artifacts/demo_cases/`; LIVE/REPLAY toggle covers this too |
| "Simulation mode" perceived as weakness by a judge | low-medium | medium | Lead with the safety rationale: *"We do not connect a personal Coinbase account in a public demo."* The SIMULATION MODE pill in the UI is the transparency badge. Mention the `Settler` interface swap point. |
| Team scope creep — directive's 5 new docs become a sink | high | medium | Defer all docs to T+5–T+6 except `BRIGHTDATA_INTEGRATION.md` (running log from T+1) |
| Go ecosystem friction surprise on MCP-x402 bridge | low | medium | TS shim slots (`adapters/mcp-bridge`, `adapters/x402-signer`) reserved; activate only if a Go path is materially slower; never block on philosophy |
| Multi-track claim looks forced | low | medium | Anchor T-Agent first; mention T-Commerce + T-Web only where the architecture genuinely overlaps |
| Judging weights differ from assumed criteria | low | low | Pivot deck order on T+5 if kickoff reveals different criteria; underlying product wins on any common rubric |
| Onsite registration / team-object trip-up | low | high | Verify lablab team-object on T+0; even solo entries need a team |
| Prompt injection inside scraped seller pages | medium | medium | Sanitize page text before LLM ingestion; surface this as a slide ("we already thought of that") |
| Solo execution risk (David alone with Claude) | medium | high | Recruit 1 frontend / x402 helper if onsite network surfaces one; otherwise descope UI fanciness, not the gate |

---

## 13. Post-hackathon path — converting the win

⊢ Bright Data Startup Program: AI-native startups, especially funded pre-seed to Series A up to $20M; bootstrapped startups may qualify for a limited amount. Position GEM² accordingly.

```
H2_Plan ≜ [
  days_1_to_14:
    apply to Bright Data Startup Program with LedgerLens + GEM² as the product line,
    swap SimulatedSettler → CoinbaseX402Settler behind the existing Settler interface
      (Base Sepolia first, tiny amounts; mainnet later under proper key custody),
    open-source the L0/L1/L2 layers under MIT in the hackathon repo,
    keep the L3 gate engine + commercial policy DSL in a separate private repo (GEM² core/proprietary),

  days_15_to_45:
    five enterprise pilots with CFOs / Heads-of-AI,
    success criterion = one paid pilot,
    publish a public audit-bundle gallery for trust-building,

  days_46_to_90:
    pre-seed raise against "trust layer for agentic commerce" thesis,
    LedgerLens v1.0: multi-chain real settlement + multi-LLM verifier ensemble + policy DSL,
    GEM² announces public Trust Gate API
]
```

---

## 14. Resolved decisions (David, 2026-05-23)

| # | Question | Resolution |
|---|---|---|
| Q1 | Onsite vs online? | **Onsite recommended** — SF Web Data Loft May 30–31; network ROI outweighs 2 days of build loss |
| Q2 | Team composition? | **Solo possible; frontend / x402 helper would be best.** Recruit if onsite network surfaces one |
| Q3 | Brand name? | **LedgerLens** (locked) |
| Q4 | Settlement layer? | **x402 protocol SIMULATION** for the public demo (no Coinbase account, no private keys, no testnet funds). The hackathon repo ships ONLY `SimulatedSettler`. Real x402 (Base Sepolia → mainnet) is a post-hackathon swap behind the `Settler` interface. (v2.2 lock — supersedes earlier "Base Sepolia testnet" decision.) |
| Q5 | OSS posture? | **Hackathon repo: full MIT. GEM² core / proprietary L3 engine: separate private repo.** Demarcate clearly in `THIRD_PARTY_NOTICES.md` |
| Stack | Language? | **Go-first** (one binary) + Next.js UI + TS shim slots reserved (likely unused) |
| Vocab | Third positioning vector? | **Trust Gate / Epistemic Settlement Gate.** "Compliance" used only for the thin compliance surface |
| EEF | Claim tag count? | **Canonical 4 in API and audit bundle.** "Speculative" is a UI label for ⊬-without-basis |
| Agents | 4 processes? | **No.** 4 ROLES (buyer / seller / verifier / gate) implemented as Go contracts inside one binary |

---

## 15. Language discipline — use / avoid

### Use (judge-facing copy)
- Agent-to-Agent Payments
- Web Data Commerce
- GEM² Trust Gate · Epistemic Settlement Gate · Audit & Trust Gate
- Trust-Gated Settlement
- Evidence-Backed Agentic Commerce
- GEM² Architectural Stack (Evidence → Memory → Verification → Release)
- L1 P-check + L2 O-check via `gem2-tpmn-checker.fly.dev` (production audit gate)
- x402-compatible payment lifecycle (simulation mode for public demo)
- "We simulate settlement, but the trust gate is real."
- **No grounded claim, no payment.**

### Avoid (in headlines or deck — fine as sub-bullets)
- "Track 3-only"
- "Security & Compliance-only project"
- "Vendor risk dashboard" / "regulatory monitor" / "threat intelligence pipeline"
- "Generic AI agent" / "generic scraper" / "RAG chatbot"
- "Compliance Gate" as the headline (use only as one of five surfaces in §5.4)

---

## 16. Sources verified 2026-05-23

**Hackathon and ecosystem**
- Bright Data Web Data UNLOCKED hackathon (lablab listing — direct fetch 403; cross-verified via search)
- Bright Data MCP repo: `github.com/brightdata/brightdata-mcp`
- Bright Data MCP docs: `docs.brightdata.com/ai/mcp-server/overview`
- AWS Bedrock AgentCore Payments (2026-05) announcement
- Stripe x402 integration (2026-02)
- Morningstar / PRNewswire — x402 SF hackathon (parallel ecosystem event)
- DEV Community — Bright Data Real-Time AI Agents Challenge (prior precedent, $3K prize)
- ScrapeAlchemist `brightdata-hack-pack` — starter kit

**x402 protocol**
- Coinbase x402 official repo + Go directory: `github.com/coinbase/x402` and `github.com/coinbase/x402/tree/main/go`
- Coinbase x402 documentation: `docs.cdp.coinbase.com/x402/welcome` + network-support page
- Coinbase + Cloudflare x402 Foundation announcement (Cloudflare blog)
- mark3labs/x402-go (community Go SDK, additional reference)
- mark3labs/mcp-go (1,880 importers — production Go MCP SDK)
- mark3labs/mcp-go-x402 (x402 transport for MCP-Go — direct LedgerLens pattern)

**GEM² production audit infrastructure (internal — GEM².AI-operated)**
- the GEM² audit-gate API spec (proprietary internal documentation) v1.1 (2026-05-18) — production audit-gate API spec, RAG-augmented, hash-only audit trail
- `Hackathon/TechEx/` — working Go integration of the audit gate (Track 1, TechEx 2026); live at `https://techex-track1.gemsquared.ai/`
- TechEx reusables (all MIT): `console/audit_gate_client.go`, `console/audit_log.go`, `console/lobstertrap.go`, `console/ce_contract_parser.go`
- Production audit gate base URL: `https://gem2-tpmn-checker.fly.dev` (staging fallback available)
- Existing report: `Docs/deep-research-report.md` — partially superseded; Bright Data routing + licensing care still hold

---

*⊢ This document is v2.0 of the LedgerLens winning strategy — the architecture deliverable from Alchy, with Kritik corrections absorbed.*
*Implementation belongs to Gineer; live-claim verification belongs to Kritik; market-reality sharpening belongs to D-Mark on explicit-call. Route accordingly.*
*Next step: `/plan-work` decomposes this into a Work-Plan with ≤9 unit-works, each carrying a CONTRACT.*
