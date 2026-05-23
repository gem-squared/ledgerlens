package auditgate

import (
	"reflect"
	"testing"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// fixtureReasons is the actual reasons array returned by gem2-tpmn-checker
// on 2026-05-23 during Unit 3 preflight (see WP-ST-1 Unit 3 result notes).
// Captured live; used to test the parser deterministically.
var fixtureReasons = []string{
	"[TYPE] I fails structural match with A: natural language string 'transfer $100 from account A to account B in USD' does not conform to TransferRequest{from, to, amount, currency} record structure",
	"[RULE-1] amount must be > 0 — FAIL: cannot extract numeric amount field from unstructured input",
	"[RULE-2] currency must be ISO4217 — FAIL: cannot extract currency field from unstructured input, though 'USD' appears to be valid ISO4217",
	"[SPT-Δe→∫de] inferring structured fields from single natural language string constitutes thin evidence for formal record requirements",
	"[EEF-⊬] extrapolating that 'account A' and 'account B' represent valid account identifiers without formal validation",
	"[DIM-grounding] 20/100 — natural language input lacks grounded structure",
	"[DIM-evidence] 0/100 — no evidence provided to validate account existence or format",
	"[DIM-logical] 40/100 — intent is clear but structure misaligned",
	"[DIM-attribution] 0/100 — no source attribution for account identifiers",
	"[DIM-temporal] 60/100 — request appears current",
	"[DIM-scope] 30/100 — scope unclear due to unstructured format",
}

func TestParseReasons_Fixture(t *testing.T) {
	parsed := ParseReasons(fixtureReasons)

	if parsed.TypeFinding == "" {
		t.Errorf("TypeFinding not extracted")
	}
	if len(parsed.Rules) != 2 {
		t.Fatalf("rules: got %d want 2", len(parsed.Rules))
	}
	if parsed.Rules[0].Index != 1 || parsed.Rules[0].Verdict != "FAIL" {
		t.Errorf("Rules[0]: %+v", parsed.Rules[0])
	}
	if parsed.Rules[0].Rule != "amount must be > 0" {
		t.Errorf("Rules[0].Rule: %q", parsed.Rules[0].Rule)
	}
	if len(parsed.SPTCodes) != 1 || parsed.SPTCodes[0] != "Δe→∫de" {
		t.Errorf("SPT codes: %+v", parsed.SPTCodes)
	}
	if len(parsed.EEFFlags) != 1 || parsed.EEFFlags[0].Tag != "⊬" {
		t.Errorf("EEF flags: %+v", parsed.EEFFlags)
	}
	if len(parsed.Dimensions) != 6 {
		t.Errorf("Dimensions: got %d want 6", len(parsed.Dimensions))
	}
	// spot-check grounding
	if parsed.Dimensions[0].Name != "grounding" || parsed.Dimensions[0].Score != 20 {
		t.Errorf("DIM[0]: %+v", parsed.Dimensions[0])
	}
}

func TestToClaimAssessments_Fixture(t *testing.T) {
	assessments := ToClaimAssessments(fixtureReasons, 25)
	if len(assessments) != 2 {
		t.Fatalf("expected 2 assessments, got %d", len(assessments))
	}
	for _, ca := range assessments {
		if ca.Status != schemas.ClaimExtrapolated {
			t.Errorf("%s: status=%q want extrapolated (⊬) — FAIL rule with stated reason", ca.ClaimID, ca.Status)
		}
		if ca.Basis == "" {
			t.Errorf("%s: basis empty (⊬ must carry basis)", ca.ClaimID)
		}
		if !reflect.DeepEqual(ca.SPTViolations, []string{"Δe→∫de"}) {
			t.Errorf("%s: SPT=%v want [Δe→∫de]", ca.ClaimID, ca.SPTViolations)
		}
		if ca.Confidence < 0.24 || ca.Confidence > 0.26 {
			t.Errorf("%s: confidence=%v want ~0.25", ca.ClaimID, ca.Confidence)
		}
	}
	// canonical ClaimID format
	if assessments[0].ClaimID != "RULE-1" {
		t.Errorf("ClaimID[0]=%q want RULE-1", assessments[0].ClaimID)
	}
}
