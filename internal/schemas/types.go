// Package schemas is the canonical source of truth for LedgerLens data shapes.
//
// Source of truth for: Docs/Bright-Data-winning-strategy.md §6.5 (v2.3).
// TS types are generated from this file via `make schemas` (tygo) into
// packages/contracts-ts/types.ts; do not hand-edit the TS.
package schemas

import "context"

// ─── Bright Data evidence ───────────────────────────────────────────────────

type SellerOffer struct {
	OfferID   string  `json:"offerId"`
	SellerID  string  `json:"sellerId"`
	Claim     string  `json:"claim"`
	PriceUSDC float64 `json:"priceUSDC"`
	SourceURL string  `json:"sourceUrl,omitempty"`
	CreatedAt string  `json:"createdAt"`
}

type BuyerRequest struct {
	RequestID    string        `json:"requestId"`
	BuyerID      string        `json:"buyerId"`
	Query        string        `json:"query"`
	MaxSpendUSDC float64       `json:"maxSpendUSDC"`
	Policy       PaymentPolicy `json:"policy"`
}

type PaymentPolicy struct {
	SpendCap              float64 `json:"spendCap"`
	ClaimGroundedRequired bool    `json:"claimGroundedRequired"`
	PublicOnly            bool    `json:"publicOnly"`
}

type EvidenceReceipt struct {
	ReceiptID         string `json:"receiptId"`
	URL               string `json:"url"`
	BrightDataProduct string `json:"brightDataProduct"` // SERP|UNLOCKER|BROWSER|SCRAPER|MCP
	FetchedAt         string `json:"fetchedAt"`
	ContentHash       string `json:"contentHash"`
	RawRef            string `json:"rawRef"`
}

// ─── GEM² claim assessment (canonical EEF — 4 tags) ─────────────────────────

type ClaimStatus string

const (
	ClaimGrounded     ClaimStatus = "grounded"     // ⊢
	ClaimInferred     ClaimStatus = "inferred"     // ⊨
	ClaimExtrapolated ClaimStatus = "extrapolated" // ⊬ (UI renders as "Speculative" when Basis is empty)
	ClaimUnknown      ClaimStatus = "unknown"      // ⊥
)

type ClaimAssessment struct {
	ClaimID       string      `json:"claimId"`
	Claim         string      `json:"claim"`
	Status        ClaimStatus `json:"status"`
	Basis         string      `json:"basis,omitempty"` // required when Status == ClaimExtrapolated
	EvidenceRefs  []string    `json:"evidenceRefs"`
	SPTViolations []string    `json:"sptViolations"` // "S->T" | "L->G" | "delta_e->int_de"
	Confidence    float64     `json:"confidence"`
}

// ─── L3 release verdict + decision packet ───────────────────────────────────

type GateVerdict string

const (
	GateApprovedByTrustGate GateVerdict = "APPROVED_BY_TRUST_GATE"
	GateBlockedByTrustGate  GateVerdict = "BLOCKED_BY_TRUST_GATE"
	GateEscalatedToHuman    GateVerdict = "ESCALATED_TO_HUMAN"
)

type DecisionPacket struct {
	DecisionID       string            `json:"decisionId"`
	RequestID        string            `json:"requestId"`
	OfferID          string            `json:"offerId"`
	Verdict          GateVerdict       `json:"verdict"`
	Reason           string            `json:"reason"`
	ClaimAssessments []ClaimAssessment `json:"claimAssessments"`
	PaymentAllowed   bool              `json:"paymentAllowed"`
	AuditBundleRef   string            `json:"auditBundleRef"`
	L1ResultID       string            `json:"l1ResultId,omitempty"` // gem2-tpmn-checker upstream result_id
	L2ResultID       string            `json:"l2ResultId,omitempty"`
}

// ─── x402 SIMULATION mode (v2.2 lock) ───────────────────────────────────────
// No chain, no private keys, no real funds. The Settler interface is the
// swap point for a post-hackathon CoinbaseX402Settler implementation.

type PaymentState string

const (
	PaymentRequired     PaymentState = "PAYMENT_REQUIRED"
	PendingVerification PaymentState = "PENDING_VERIFICATION"
	ApprovedByTrustGate PaymentState = "APPROVED_BY_TRUST_GATE"
	BlockedByTrustGate  PaymentState = "BLOCKED_BY_TRUST_GATE"
	EscalatedToHuman    PaymentState = "ESCALATED_TO_HUMAN"
	SimulatedSettled    PaymentState = "SIMULATED_SETTLED"
)

type SettlementMode string

const (
	ModeSimulation SettlementMode = "simulation" // hackathon default — the only mode shipped
	ModeMainnet    SettlementMode = "mainnet"    // post-hackathon — not used in this repo
	ModeTestnet    SettlementMode = "testnet"    // post-hackathon — not used in this repo
)

type SimulatedSettlement struct {
	SettlementID    string         `json:"settlementId,omitempty"` // null when blocked
	DecisionID      string         `json:"decisionId"`
	Mode            SettlementMode `json:"mode"`            // always "simulation" in hackathon repo
	Network         string         `json:"network"`         // always "demo-local"
	Asset           string         `json:"asset"`           // always "USDC-demo"
	Status          PaymentState   `json:"status"`
	AmountUSDC      float64        `json:"amountUSDC"`      // 0 when blocked
	Reason          string         `json:"reason"`
	RealTransaction bool           `json:"realTransaction"` // always false
	PrivateKeysUsed bool           `json:"privateKeysUsed"` // always false
	RealFundsUsed   bool           `json:"realFundsUsed"`   // always false
	Ts              string         `json:"ts"`              // ISO8601 UTC
}

// Settler is the interface that a real Coinbase x402 implementation would
// satisfy post-hackathon. The hackathon repo ships ONLY SimulatedSettler.
type Settler interface {
	Settle(ctx context.Context, decision DecisionPacket) (SimulatedSettlement, error)
}
