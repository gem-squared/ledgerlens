// Package evidence converts the EvidenceReceipt[] emitted by
// internal/brightdata into the `evidence: []string` payload the gem2-tpmn-checker
// audit gate expects (per AUDIT_GATE_API.md v1.1 RAG augmentation pattern).
//
// Each chunk is a self-describing line of header metadata + a bounded body
// excerpt so the LLM-judge can ground individual claims in named evidence.
package evidence

import (
	"fmt"
	"os"
	"strings"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// MaxBodyChars caps each evidence chunk's body excerpt to stay within
// reasonable token budgets. The audit gate's LLM call cost scales with input,
// so we trim. Set generously enough that the LLM still has signal.
const MaxBodyChars = 1500

// WrapReceipts returns evidence chunks ready for `PCheckRequest.Evidence`
// or `OCheckRequest.Evidence`. Each chunk has the shape:
//
//	[receipt <id> | product=<P> | url=<url> | fetched=<ts> | hash=<sha256>]
//	<body excerpt, max MaxBodyChars chars>
//
// If a receipt's RawRef file is missing, the chunk still includes the
// metadata line (header-only). Reading errors are logged via the returned
// error slice but do not abort the wrap — missing one chunk should not
// kill the whole verification.
func WrapReceipts(receipts []schemas.EvidenceReceipt) (chunks []string, readErrors []error) {
	for _, r := range receipts {
		header := fmt.Sprintf(
			"[receipt %s | product=%s | url=%s | fetched=%s | hash=%s]",
			r.ReceiptID, r.BrightDataProduct, r.URL, r.FetchedAt, r.ContentHash,
		)
		body := ""
		if r.RawRef != "" {
			raw, err := os.ReadFile(r.RawRef)
			if err != nil {
				readErrors = append(readErrors, fmt.Errorf("evidence: read %s: %w", r.RawRef, err))
			} else {
				body = trim(string(raw), MaxBodyChars)
			}
		}
		if body == "" {
			chunks = append(chunks, header)
		} else {
			chunks = append(chunks, header+"\n"+body)
		}
	}
	return chunks, readErrors
}

func trim(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "...(truncated)"
}

// JoinHashable returns a deterministic concatenation of chunks suitable for
// hashing into a single evidence_hash. Order-sensitive — call WrapReceipts
// in a stable order before this. Used for local audit-trail correlation.
func JoinHashable(chunks []string) string {
	return strings.Join(chunks, "\n---\n")
}
