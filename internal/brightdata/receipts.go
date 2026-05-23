package brightdata

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// ReceiptStore persists EvidenceReceipt records to artifacts/fetch_receipts/.
// Layout:
//
//	<root>/receipts.jsonl          (append-only)
//	<root>/<receiptId>.<ext>       (raw body — referenced by receipt.RawRef)
//
// The JSONL is the audit trail; the per-receipt files are the raw evidence.
type ReceiptStore struct {
	root string // e.g. "artifacts/fetch_receipts"
}

func NewReceiptStore(root string) (*ReceiptStore, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("receipt store: mkdir %s: %w", root, err)
	}
	return &ReceiptStore{root: root}, nil
}

// Write persists one receipt + its raw body. The body extension is chosen
// from the product (HTML for unlocker/browser, JSON for SERP, TXT for MCP).
func (s *ReceiptStore) Write(receipt schemas.EvidenceReceipt, body []byte, ext string) (schemas.EvidenceReceipt, error) {
	if receipt.ReceiptID == "" {
		receipt.ReceiptID = "ev_" + randID(8)
	}
	if receipt.FetchedAt == "" {
		receipt.FetchedAt = time.Now().UTC().Format(time.RFC3339)
	}
	if receipt.ContentHash == "" {
		sum := sha256.Sum256(body)
		receipt.ContentHash = "sha256:" + hex.EncodeToString(sum[:])
	}

	rawPath := filepath.Join(s.root, receipt.ReceiptID+"."+ext)
	if err := os.WriteFile(rawPath, body, 0o644); err != nil {
		return receipt, fmt.Errorf("receipt store: write raw %s: %w", rawPath, err)
	}
	receipt.RawRef = rawPath

	jsonl := filepath.Join(s.root, "receipts.jsonl")
	f, err := os.OpenFile(jsonl, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return receipt, fmt.Errorf("receipt store: open jsonl: %w", err)
	}
	defer f.Close()
	line, err := json.Marshal(receipt)
	if err != nil {
		return receipt, fmt.Errorf("receipt store: marshal: %w", err)
	}
	if _, err := f.Write(append(line, '\n')); err != nil {
		return receipt, fmt.Errorf("receipt store: append: %w", err)
	}
	return receipt, nil
}

func randID(nBytes int) string {
	b := make([]byte, nBytes)
	if _, err := rand.Read(b); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
