package api

import (
	"encoding/json"
	"errors"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gem-squared/ledgerlens/internal/paymentgate"
	"github.com/gem-squared/ledgerlens/internal/schemas"

	"github.com/gin-gonic/gin"
)

// Stats is the aggregate metrics payload for the Verification Infrastructure
// Dashboard. Every value is derived from on-disk audit bundles — no separate
// analytics store. Sample size scales with the directory; ~10ms scan up to
// several hundred bundles.
type Stats struct {
	DealsAudited                int     `json:"dealsAudited"`
	Approved                    int     `json:"approved"`
	Blocked                     int     `json:"blocked"`
	EscalatedToHuman            int     `json:"escalatedToHuman"`
	AvgAuditScore               int     `json:"avgAuditScore"`
	AvgVerificationLatencyMs    int64   `json:"avgVerificationLatencyMs"`
	SimulatedSpendPreventedUSDC float64 `json:"simulatedSpendPreventedUSDC"`
	BrightDataReceipts          int     `json:"brightDataReceipts"`
	AuditBundlesExported        int     `json:"auditBundlesExported"`
	SampleSize                  int     `json:"sampleSize"`
	LastUpdatedAt               string  `json:"lastUpdatedAt,omitempty"`
	AuditGateURL                string  `json:"auditGateUrl"`
	AuditGateAvgLatencyMs       int64   `json:"auditGateAvgLatencyMs"`
	ModesBreakdown              Modes   `json:"modesBreakdown"`
}

type Modes struct {
	Live      int `json:"live"`
	PreWarmed int `json:"preWarmed"`
	Replay    int `json:"replay"`
	Unknown   int `json:"unknown"`
}

// BundleSummary is the list-level shape returned by GET /api/v1/audit-bundles.
// Excludes the heavy reasons[] / evidenceReceipts[] payloads — just enough
// for Recent Activity rows + "View" link.
type BundleSummary struct {
	DecisionID       string  `json:"decisionId"`
	BundleID         string  `json:"bundleId"`
	Verdict          string  `json:"verdict"`
	Mode             string  `json:"mode"`
	Query            string  `json:"query"`
	DurationMs       int64   `json:"durationMs"`
	L1Score          int     `json:"l1Score"`
	L2Score          int     `json:"l2Score,omitempty"`
	L2Skipped        bool    `json:"l2Skipped"`
	PaymentAllowed   bool    `json:"paymentAllowed"`
	SettlementStatus string  `json:"settlementStatus,omitempty"`
	SettlementID     string  `json:"settlementId,omitempty"`
	AmountUSDC       float64 `json:"amountUSDC"`
	CreatedAt        string  `json:"createdAt"`
	EvidenceCount    int     `json:"evidenceCount"`
}

const auditGateURL = "gem2-tpmn-checker.fly.dev"

// getStats serves GET /api/v1/stats. Cache-Control: no-store so the
// dashboard's post-run refetch always sees fresh counter values.
func (s *Server) getStats(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0, must-revalidate")
	stats, err := s.computeStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, stats)
}

// listAuditBundles serves GET /api/v1/audit-bundles (newest first, capped 50).
func (s *Server) listAuditBundles(c *gin.Context) {
	c.Header("Cache-Control", "no-store, max-age=0, must-revalidate")
	out, err := s.loadBundleSummaries(50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"bundles": out})
}

// ─── computeStats ─────────────────────────────────────────────────────────

func (s *Server) computeStats() (Stats, error) {
	stats := Stats{AuditGateURL: auditGateURL}

	bundles, err := s.scanAllBundles()
	if err != nil {
		return stats, err
	}

	var sumScore, scoreCount int64
	var sumDuration, durationCount int64
	var sumGateLatency, gateLatencyCount int64
	var sumBlockedAmount float64
	var lastTs string

	for _, b := range bundles {
		stats.DealsAudited++
		stats.BrightDataReceipts += len(b.EvidenceReceipts)

		// Verdict bucket
		switch b.Decision.Verdict {
		case schemas.GateApprovedByTrustGate:
			stats.Approved++
		case schemas.GateBlockedByTrustGate:
			stats.Blocked++
			// "Simulated Spend Prevented" — the buyer would have paid this on
			// the proposed offer, but the gate blocked. Sum these amounts.
			sumBlockedAmount += b.SellerOffer.PriceUSDC
		case schemas.GateEscalatedToHuman:
			stats.EscalatedToHuman++
		}

		// Audit score (composite — average of L1 and L2 scores when both ran)
		if b.L1 != nil {
			sumScore += int64(b.L1.Score)
			scoreCount++
			sumGateLatency += b.L1.Meta.DurationMs
			gateLatencyCount++
		}
		if b.L2 != nil {
			sumScore += int64(b.L2.Score)
			scoreCount++
			sumGateLatency += b.L2.Meta.DurationMs
			gateLatencyCount++
		}

		// End-to-end verification latency from bundle timestamps
		if b.Timestamps.StartedAt != "" && b.Timestamps.FinishedAt != "" {
			start, e1 := time.Parse(time.RFC3339, b.Timestamps.StartedAt)
			end, e2 := time.Parse(time.RFC3339, b.Timestamps.FinishedAt)
			if e1 == nil && e2 == nil && end.After(start) {
				sumDuration += end.Sub(start).Milliseconds()
				durationCount++
			}
		}

		if b.Timestamps.FinishedAt > lastTs {
			lastTs = b.Timestamps.FinishedAt
		}

		// Mode breakdown
		switch modeFromRequestID(b.BuyerRequest.RequestID) {
		case "live":
			stats.ModesBreakdown.Live++
		case "prewarmed":
			stats.ModesBreakdown.PreWarmed++
		case "replay":
			stats.ModesBreakdown.Replay++
		default:
			stats.ModesBreakdown.Unknown++
		}
	}

	stats.AuditBundlesExported = stats.DealsAudited
	stats.SampleSize = stats.DealsAudited
	stats.SimulatedSpendPreventedUSDC = round4(sumBlockedAmount)
	stats.LastUpdatedAt = lastTs
	if scoreCount > 0 {
		stats.AvgAuditScore = int(sumScore / scoreCount)
	}
	if durationCount > 0 {
		stats.AvgVerificationLatencyMs = sumDuration / durationCount
	}
	if gateLatencyCount > 0 {
		stats.AuditGateAvgLatencyMs = sumGateLatency / gateLatencyCount
	}
	return stats, nil
}

// ─── Recent Activity (bundles list) ──────────────────────────────────────

func (s *Server) loadBundleSummaries(limit int) ([]BundleSummary, error) {
	bundles, err := s.scanAllBundles()
	if err != nil {
		return nil, err
	}
	sort.Slice(bundles, func(i, j int) bool {
		// Newest first
		return bundles[i].Timestamps.FinishedAt > bundles[j].Timestamps.FinishedAt
	})
	if limit > 0 && len(bundles) > limit {
		bundles = bundles[:limit]
	}
	out := make([]BundleSummary, 0, len(bundles))
	for _, b := range bundles {
		out = append(out, summarize(b))
	}
	return out, nil
}

func summarize(b paymentgate.AuditBundle) BundleSummary {
	var durationMs int64
	if b.Timestamps.StartedAt != "" && b.Timestamps.FinishedAt != "" {
		start, e1 := time.Parse(time.RFC3339, b.Timestamps.StartedAt)
		end, e2 := time.Parse(time.RFC3339, b.Timestamps.FinishedAt)
		if e1 == nil && e2 == nil && end.After(start) {
			durationMs = end.Sub(start).Milliseconds()
		}
	}
	l1Score, l2Score := 0, 0
	l2Skipped := true
	if b.L1 != nil {
		l1Score = b.L1.Score
	}
	if b.L2 != nil {
		l2Score = b.L2.Score
		l2Skipped = false
	}
	return BundleSummary{
		DecisionID:       b.DecisionID,
		BundleID:         b.BundleID,
		Verdict:          string(b.Decision.Verdict),
		Mode:             modeFromRequestID(b.BuyerRequest.RequestID),
		Query:            b.BuyerRequest.Query,
		DurationMs:       durationMs,
		L1Score:          l1Score,
		L2Score:          l2Score,
		L2Skipped:        l2Skipped,
		PaymentAllowed:   b.Decision.PaymentAllowed,
		SettlementStatus: string(b.Settlement.Status),
		SettlementID:     b.Settlement.SettlementID,
		AmountUSDC:       b.Settlement.AmountUSDC,
		CreatedAt:        b.Timestamps.FinishedAt,
		EvidenceCount:    len(b.EvidenceReceipts),
	}
}

// ─── Shared bundle scan ──────────────────────────────────────────────────

func (s *Server) scanAllBundles() ([]paymentgate.AuditBundle, error) {
	pattern := filepath.Join(s.BundlesDir, "*.json")
	files, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}
	if len(files) == 0 {
		// Not an error — just no data yet.
		if errors.Is(err, fs.ErrNotExist) {
			return nil, nil
		}
	}
	out := make([]paymentgate.AuditBundle, 0, len(files))
	for _, fpath := range files {
		raw, err := os.ReadFile(fpath)
		if err != nil {
			continue
		}
		var b paymentgate.AuditBundle
		if err := json.Unmarshal(raw, &b); err != nil {
			continue
		}
		out = append(out, b)
	}
	return out, nil
}

// modeFromRequestID derives the run mode from the BuyerRequest.RequestID prefix.
// (Bundle doesn't store mode directly yet; conventional prefix is reliable.)
func modeFromRequestID(id string) string {
	switch {
	case strings.HasPrefix(id, "req_deal_"):
		return "live"
	case strings.HasPrefix(id, "req_prewarmed_"):
		return "prewarmed"
	case strings.HasPrefix(id, "req_caseA") || strings.HasPrefix(id, "req_caseB"):
		return "replay"
	default:
		return "unknown"
	}
}

func round4(f float64) float64 {
	// 4 decimal places — appropriate for USDC micropayments.
	return float64(int64(f*1e4+0.5)) / 1e4
}
