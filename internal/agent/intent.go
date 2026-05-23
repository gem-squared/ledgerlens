package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

// BuyerIntent is the structured interpretation of a judge's natural-language
// data-purchase request. Produced by Extract from the raw `query` string.
//
// If OffDomain is true, the request is not a data-purchase ask — the UI
// should show PoliteReject instead of running the pipeline.
type BuyerIntent struct {
	OffDomain          bool     `json:"offDomain"`
	PoliteReject       string   `json:"politeReject,omitempty"`
	DataNeed           string   `json:"dataNeed,omitempty"`
	Freshness          string   `json:"freshness,omitempty"`
	MaxSpendUSDC       float64  `json:"maxSpendUSDC,omitempty"`
	PolicyRequirements []string `json:"policyRequirements,omitempty"`
	SearchTerms        []string `json:"searchTerms,omitempty"`
}

const intentSystem = `You are the Buyer Agent in LedgerLens — a system that verifies seller claims with public web evidence before authorizing autonomous payment.

Your job: interpret a natural-language data-purchase request and extract a structured intent.

Domain scope: requests to buy live web data, an API feed, a data service, or a similar information product.

Off-domain examples (REJECT politely):
- "buy me a pokemon card"
- "transfer money to alice"
- "summarize this PDF"
- "book a flight"

Respond with valid JSON ONLY. No prose. No markdown fences.`

const intentUserTmpl = `USER REQUEST:
%q

If the request is OFF-DOMAIN, respond with EXACTLY:
{"offDomain":true,"politeReject":"This demo is scoped to autonomous web-data purchase requests. Try the default NYSE/NASDAQ market data request."}

Otherwise respond with:
{
  "offDomain": false,
  "dataNeed": "<one sentence — what is the buyer looking for>",
  "freshness": "<required freshness, e.g. 'real-time', '1-second', 'daily', 'historical'>",
  "maxSpendUSDC": <number — extract from request if stated, else a reasonable default like 0.001>,
  "policyRequirements": ["<requirement>", "..."],
  "searchTerms": ["<3-5 web search terms to discover candidate providers>"]
}

JSON ONLY:`

// Extract calls Anthropic Haiku to interpret a judge's request into a
// BuyerIntent. Returns an OffDomain intent + polite reject if the request
// is not a data-purchase ask.
func Extract(ctx context.Context, query, anthropicAPIKey string) (BuyerIntent, error) {
	if query == "" {
		return BuyerIntent{}, fmt.Errorf("intent: empty query")
	}

	user := fmt.Sprintf(intentUserTmpl, query)
	text, _, err := callAnthropic(ctx, anthropicAPIKey, ModelHaiku, intentSystem, user, 800)
	if err != nil {
		return BuyerIntent{}, fmt.Errorf("intent: %w", err)
	}

	cleaned := extractJSON(text)
	var out BuyerIntent
	if err := json.Unmarshal([]byte(cleaned), &out); err != nil {
		return BuyerIntent{}, fmt.Errorf("intent: parse %q: %w", snippet([]byte(cleaned)), err)
	}

	// Sanity defaults
	if !out.OffDomain && out.MaxSpendUSDC == 0 {
		out.MaxSpendUSDC = 0.001
	}
	return out, nil
}
