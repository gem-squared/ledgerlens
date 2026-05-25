package agent

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/evidence"
)

// Synthesize uses Anthropic Haiku to construct a plausible SellerOffer from
// the buyer's intent + the evidence chunks Bright Data discovered. The LLM
// reads the actual web content and infers what claim a real seller WOULD
// make based on what's actually advertised in the evidence.
//
// IMPORTANT: the synthesizer is instructed to faithfully reflect the
// evidence — including any limitations or staleness. The Trust Gate is the
// one that judges grounding; the synthesizer should not pre-judge.
func Synthesize(ctx context.Context, intent BuyerIntent, receipts []schemas.EvidenceReceipt, anthropicAPIKey string) (schemas.SellerOffer, error) {
	if len(receipts) == 0 {
		return schemas.SellerOffer{}, fmt.Errorf("synthesize: no evidence receipts")
	}

	chunks, _ := evidence.WrapReceipts(receipts)
	if len(chunks) == 0 {
		return schemas.SellerOffer{}, fmt.Errorf("synthesize: no evidence chunks after wrap")
	}

	intentJSON, _ := json.Marshal(intent)

	user := fmt.Sprintf(synthesizeUserTmpl, string(intentJSON), strings.Join(chunks, "\n---\n"))
	text, _, err := callAnthropic(ctx, anthropicAPIKey, ModelHaiku, synthesizeSystem, user, 800)
	if err != nil {
		return schemas.SellerOffer{}, fmt.Errorf("synthesize: %w", err)
	}

	cleaned := extractJSON(text)
	var parsed struct {
		Claim     string  `json:"claim"`
		PriceUSDC float64 `json:"priceUSDC"`
		SourceURL string  `json:"sourceUrl"`
	}
	if err := json.Unmarshal([]byte(cleaned), &parsed); err != nil {
		return schemas.SellerOffer{}, fmt.Errorf("synthesize: parse %q: %w", snippet([]byte(cleaned)), err)
	}
	if parsed.Claim == "" {
		return schemas.SellerOffer{}, fmt.Errorf("synthesize: empty claim from model")
	}
	if parsed.SourceURL == "" {
		// Fall back to the first receipt's URL.
		parsed.SourceURL = receipts[0].URL
	}
	if parsed.PriceUSDC <= 0 {
		// Half of buyer's max spend, or a tiny default.
		if intent.MaxSpendUSDC > 0 {
			parsed.PriceUSDC = intent.MaxSpendUSDC / 2
		} else {
			parsed.PriceUSDC = 0.001
		}
	}

	return schemas.SellerOffer{
		OfferID:   "offer_" + randHex(8),
		SellerID:  "seller_" + hashShort(parsed.SourceURL),
		Claim:     parsed.Claim,
		PriceUSDC: parsed.PriceUSDC,
		SourceURL: parsed.SourceURL,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}, nil
}

const synthesizeSystem = `You are the Seller Offer Agent in LedgerLens.

You read public-web evidence collected by Bright Data and synthesize what a real data-provider would claim about their service, based on what's actually advertised in the evidence.

Be FAITHFUL to the evidence:
- If the evidence shows current, comprehensive data → your claim reflects that.
- If the evidence is stale, partial, or thin → your claim must reflect that limitation, not paper over it.
- Do not invent capabilities the evidence does not mention.

If MULTIPLE candidates appear in the evidence, choose the SINGLE most promising one for the buyer's intent and return EXACTLY ONE offer for it. Do not return arrays. Do not return multiple objects.

The Trust Gate will audit the resulting claim against the evidence. Your job is fidelity, not flattery.

Respond with EXACTLY ONE JSON OBJECT — no array brackets, no multiple objects, no prose, no markdown fences.`

const synthesizeUserTmpl = `BUYER INTENT:
%s

EVIDENCE (each chunk is a header line + a public-web body excerpt):
%s

Construct EXACTLY ONE SellerOffer JSON object matching what the evidence advertises for the best candidate:
{
  "claim": "<one sentence — what the provider would claim, faithful to the evidence>",
  "priceUSDC": <number — extract from evidence if visible, else use buyer.maxSpendUSDC * 0.5>,
  "sourceUrl": "<most authoritative URL from the evidence for the chosen candidate>"
}

EXACTLY ONE OBJECT, JSON ONLY:`

func randHex(n int) string {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}

func hashShort(s string) string {
	// Strip the scheme + host for a stable ID across path variations.
	if u, err := url.Parse(s); err == nil && u.Host != "" {
		s = u.Host + u.Path
	}
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:4])
}
