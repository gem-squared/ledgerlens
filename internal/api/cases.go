// Package api exposes LedgerLens's HTTP surface to the Next.js demo UI.
package api

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// Canonical demo cases — wired into the Run handler so the UI's "Run Case A"
// and "Run Case B" buttons map to known scenarios.

// CaseDef bundles everything needed to exercise the orchestrator once.
type CaseDef struct {
	ID           string
	Title        string
	Description  string
	Buyer        schemas.BuyerRequest
	Offer        schemas.SellerOffer
	EvidenceDir  string                       // where to write evidence raw bodies
	WriteRawFn   func(dir string) []schemas.EvidenceReceipt
}

// CaseA — BLOCKED. Seller claims real-time freshness; evidence is a stale archive.
var CaseA = CaseDef{
	ID:          "a",
	Title:       "Case A — Blocked Payment",
	Description: "Seller claims 'live real-time NYSE+NASDAQ feed' but the only public evidence is a 6-month-old archive snapshot. L1 P-check denies; no settlement fires.",
	Buyer: schemas.BuyerRequest{
		RequestID: "req_caseA", BuyerID: "buyer_caseA",
		Query: "real-time NYSE NASDAQ price feed", MaxSpendUSDC: 0.01,
		Policy: schemas.PaymentPolicy{
			SpendCap: 0.01, PublicOnly: true, ClaimGroundedRequired: true,
		},
	},
	Offer: schemas.SellerOffer{
		OfferID: "offer_caseA", SellerID: "seller_caseA",
		Claim:     "Live real-time NYSE+NASDAQ price feed with 1-second freshness",
		PriceUSDC: 0.001,
		SourceURL: "https://example-vendor.com/status",
		CreatedAt: "", // filled at run time
	},
	WriteRawFn: func(dir string) []schemas.EvidenceReceipt {
		return []schemas.EvidenceReceipt{
			writeEvidence(dir, "ev_caseA_1", "html",
				`<html><title>Vendor Status — archive from 2025-11-15</title>
<body>
Status page from November 15, 2025 (snapshotted). Last update 6 months ago.
This is a Wayback-style archive page — not a live feed.
Coverage shown: NYSE only. No NASDAQ. No real-time prices visible.
</body></html>`),
		}
	},
}

// CaseB — APPROVED. Seller's claim is directly supported by live evidence.
var CaseB = CaseDef{
	ID:          "b",
	Title:       "Case B — Approved Payment",
	Description: "Seller claims 'live NYSE+NASDAQ feed at $0.001/query, 1-second freshness'. Bright Data fetches a live status page that supports every subclaim. L1 + L2 both green; simulated x402 settlement fires.",
	Buyer: schemas.BuyerRequest{
		RequestID: "req_caseB", BuyerID: "buyer_caseB",
		Query: "real-time NYSE NASDAQ price feed", MaxSpendUSDC: 0.01,
		Policy: schemas.PaymentPolicy{
			SpendCap: 0.01, PublicOnly: true, ClaimGroundedRequired: true,
		},
	},
	Offer: schemas.SellerOffer{
		OfferID: "offer_caseB", SellerID: "seller_caseB",
		Claim:     "Live NYSE+NASDAQ price feed, 1-second freshness, $0.001/query",
		PriceUSDC: 0.001,
		SourceURL: "https://example-vendor.com/status",
		CreatedAt: "",
	},
	WriteRawFn: func(dir string) []schemas.EvidenceReceipt {
		return []schemas.EvidenceReceipt{
			writeEvidence(dir, "ev_caseB_1", "html", fmt.Sprintf(
				`<html><title>Vendor Status — Live</title>
<body>
Status as of %s:
Live NYSE all symbols feed: GREEN. Latency: 712 ms (last 60 s avg).
Live NASDAQ all symbols feed: GREEN. Latency: 689 ms (last 60 s avg).
Update frequency: every 1 second.
Pricing: $0.001 / query for ≤1M queries / month.
</body></html>`, time.Now().UTC().Format(time.RFC3339))),
			writeEvidence(dir, "ev_caseB_2", "json",
				`{"organic":[{"link":"https://example-vendor.com/pricing","title":"Example Vendor Pricing","description":"$0.001 per query, NYSE+NASDAQ live feeds, 1-second freshness","global_rank":1}]}`),
		}
	},
}

// AllCases returns the case definitions in canonical order.
func AllCases() []CaseDef { return []CaseDef{CaseA, CaseB} }

// CaseByID returns a case definition or nil.
func CaseByID(id string) *CaseDef {
	for _, c := range AllCases() {
		if c.ID == id {
			return &c
		}
	}
	return nil
}

// writeEvidence drops a raw evidence body and returns a populated receipt.
func writeEvidence(dir, receiptID, ext, body string) schemas.EvidenceReceipt {
	path := filepath.Join(dir, receiptID+"."+ext)
	_ = os.WriteFile(path, []byte(body), 0o644)
	return schemas.EvidenceReceipt{
		ReceiptID:         receiptID,
		URL:               "https://example-vendor.com/" + receiptID,
		BrightDataProduct: "BROWSER",
		FetchedAt:         time.Now().UTC().Format(time.RFC3339),
		ContentHash:       "sha256:" + receiptID,
		RawRef:            path,
	}
}
