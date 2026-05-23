package paymentgate

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
	"github.com/gem-squared/ledgerlens/internal/trustgate/auditgate"
)

// AuditBundle is the regulator-replay-ready record of one decision. Every
// payment (approved or blocked) produces a bundle. Bundles are hash-chained
// — each bundle records the SHA-256 of the evidence corpus that drove it,
// so a future verifier can prove WHICH evidence snapshot was consulted.
type AuditBundle struct {
	BundleID         string                       `json:"bundleId"`
	DecisionID       string                       `json:"decisionId"`
	BuyerRequest     schemas.BuyerRequest         `json:"buyerRequest"`
	SellerOffer      schemas.SellerOffer          `json:"sellerOffer"`
	EvidenceReceipts []schemas.EvidenceReceipt    `json:"evidenceReceipts"`
	EvidenceHash     string                       `json:"evidenceHash"`           // sha256 of joined chunk text
	L1               *auditgate.PCheckResponse    `json:"l1,omitempty"`
	L2               *auditgate.OCheckResponse    `json:"l2,omitempty"`
	ClaimAssessments []schemas.ClaimAssessment    `json:"claimAssessments"`
	Decision         schemas.DecisionPacket       `json:"decision"`
	Settlement       schemas.SimulatedSettlement  `json:"settlement"`
	Timestamps       BundleTimestamps             `json:"timestamps"`
}

type BundleTimestamps struct {
	StartedAt  string `json:"startedAt"`
	FinishedAt string `json:"finishedAt"`
}

// BundleStore writes audit bundles as JSON files into a root directory
// (typically artifacts/audit_bundles/). One file per decision, named by
// DecisionID for direct lookup.
type BundleStore struct {
	root string
}

func NewBundleStore(root string) (*BundleStore, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("bundle store: mkdir %s: %w", root, err)
	}
	return &BundleStore{root: root}, nil
}

// Write persists a bundle and returns the file path. The BundleID and
// EvidenceHash are populated if empty.
func (s *BundleStore) Write(b AuditBundle, evidenceChunks []string) (string, AuditBundle, error) {
	if b.BundleID == "" {
		b.BundleID = "ab_" + randHex(8)
	}
	if b.EvidenceHash == "" {
		joined := strings.Join(evidenceChunks, "\n---\n")
		sum := sha256.Sum256([]byte(joined))
		b.EvidenceHash = "sha256:" + hex.EncodeToString(sum[:])
	}
	if b.Timestamps.FinishedAt == "" {
		b.Timestamps.FinishedAt = time.Now().UTC().Format(time.RFC3339)
	}

	path := filepath.Join(s.root, b.DecisionID+".json")
	raw, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return "", b, fmt.Errorf("bundle store: marshal: %w", err)
	}
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return "", b, fmt.Errorf("bundle store: write %s: %w", path, err)
	}
	return path, b, nil
}

// Read loads a bundle by DecisionID.
func (s *BundleStore) Read(decisionID string) (*AuditBundle, error) {
	path := filepath.Join(s.root, decisionID+".json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("bundle store: read %s: %w", path, err)
	}
	var b AuditBundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return nil, fmt.Errorf("bundle store: unmarshal: %w", err)
	}
	return &b, nil
}
