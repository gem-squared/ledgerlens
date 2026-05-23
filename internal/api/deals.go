package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/gem-squared/ledgerlens/internal/agent"
	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"

	"github.com/gin-gonic/gin"
)

// DealRunRequest is the JSON body for POST /api/v1/deals/run.
type DealRunRequest struct {
	Query           string  `json:"query"           binding:"required"`
	MaxSpendUSDC    float64 `json:"maxSpendUSDC,omitempty"`
	RequireGrounded bool    `json:"requireGrounded,omitempty"`
	Mode            string  `json:"mode,omitempty"` // "live" (default) | "prewarmed" (Slice 3) | "replay" (Case A/B)
}

// DealRunResult is the JSON body returned by POST /api/v1/deals/run.
type DealRunResult struct {
	Mode             string                       `json:"mode"`
	JudgeRequest     string                       `json:"judgeRequest"`
	BuyerIntent      agent.BuyerIntent            `json:"buyerIntent"`
	AgentNarrative   []string                     `json:"agentNarrative"`
	EvidenceReceipts []schemas.EvidenceReceipt    `json:"evidenceReceipts"`
	SellerOffer      schemas.SellerOffer          `json:"sellerOffer"`
	Decision         schemas.DecisionPacket       `json:"decision"`
	Settlement       schemas.SimulatedSettlement  `json:"settlement"`
	L1               *auditgate.PCheckResponse    `json:"l1,omitempty"`
	L2               *auditgate.OCheckResponse    `json:"l2,omitempty"`
	BundlePath       string                       `json:"bundlePath"`
	FinalReport      agent.FinalReport            `json:"finalReport"`
	DurationMs       int64                        `json:"durationMs"`
}

// runDeal handles POST /api/v1/deals/run (JSON, blocking). Calls the shared
// runDealPipeline with a nil emitter. For step-by-step progress use the
// SSE variant POST /api/v1/deals/run-stream.
func (s *Server) runDeal(c *gin.Context) {
	var req DealRunRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body: " + err.Error()})
		return
	}
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}
	if req.Mode == "" {
		req.Mode = "live"
	}
	if req.Mode != "live" {
		c.JSON(http.StatusNotImplemented, gin.H{
			"error": "mode not implemented in slice 2",
			"mode":  req.Mode,
			"note":  "Slice 3 (prewarmed) — for now use mode=live",
		})
		return
	}
	if s.AnthropicAPIKey == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "ANTHROPIC_API_KEY not configured on server"})
		return
	}
	if s.SERP == nil || s.Unlocker == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Bright Data clients not configured (BRIGHTDATA_API_TOKEN + zones)"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	result, err := s.runDealPipeline(ctx, req, nil)
	if err != nil {
		// Off-domain rejection gets a distinct status code so the UI can
		// render the polite-reject banner correctly.
		if strings.HasPrefix(err.Error(), "off_domain:") {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"error":        "off_domain",
				"politeReject": strings.TrimPrefix(err.Error(), "off_domain: "),
				"judgeRequest": req.Query,
			})
			return
		}
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

// extractCandidateURLs pulls top organic links from a SERP JSON file.
// Defensive: SERP JSON shape can vary; uses string-search to be resilient.
func extractCandidateURLs(serpJSONPath string, maxN int) []string {
	body, err := readFile(serpJSONPath)
	if err != nil {
		return nil
	}
	// Find all "link":"..." substrings — Bright Data's SERP JSON uses this key
	// inside organic[]. Cheap and resilient to schema drift.
	urls := []string{}
	s := string(body)
	for {
		i := strings.Index(s, `"link":"`)
		if i < 0 {
			break
		}
		s = s[i+len(`"link":"`):]
		j := strings.Index(s, `"`)
		if j < 0 {
			break
		}
		candidate := s[:j]
		s = s[j+1:]
		// Filter out Google-internal redirects + noisy domains
		if strings.HasPrefix(candidate, "http") &&
			!strings.Contains(candidate, "google.com/url") &&
			!strings.Contains(candidate, "youtube.com/") {
			urls = append(urls, candidate)
			if len(urls) >= maxN {
				break
			}
		}
	}
	return urls
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}

func readFile(path string) ([]byte, error) {
	return readFileImpl(path)
}
