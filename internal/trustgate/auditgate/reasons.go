package auditgate

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gem-squared/ledgerlens/internal/schemas"
)

// ParsedReasons is the structured view of a gate response's flat `reasons`
// array. The audit-gate API emits bracketed tags ([TYPE], [RULE-N], [SPT-X],
// [EEF-X], [EVIDENCE-N|⊥], [DIM-X]) — we extract them here.
type ParsedReasons struct {
	TypeFinding string         // from [TYPE]
	Rules       []RuleFinding  // from [RULE-N]
	SPTCodes    []string       // from [SPT-X] (codes only, e.g. "S→T")
	EEFFlags    []EEFFlag      // from [EEF-X]
	Evidence    []EvidenceUse  // from [EVIDENCE-N|⊥]
	Dimensions  []DimensionScore // from [DIM-X]
}

type RuleFinding struct {
	Index   int    // N from RULE-N
	Rule    string // rule text (left of em dash)
	Verdict string // PASS | FAIL
	Reason  string // text after PASS|FAIL:
}

type EEFFlag struct {
	Tag  string // ⊢ ⊨ ⊬ ⊥
	Text string
}

type EvidenceUse struct {
	Index   int    // N for [EVIDENCE-N], -1 for [EVIDENCE-⊥]
	Excerpt string // full reason body
}

type DimensionScore struct {
	Name  string
	Score int
	Note  string
}

// ─── Compiled regexes ───────────────────────────────────────────────────────
//
// The audit gate uses an em dash (U+2014 "—") between rule text and the
// PASS/FAIL marker, and between dimension score and note. We accept either
// em dash or hyphen for resilience.

var (
	reType     = regexp.MustCompile(`^\[TYPE\]\s+(.+)$`)
	reRule     = regexp.MustCompile(`^\[RULE-(\d+)\]\s+(.+?)\s*[—-]\s*(PASS|FAIL):\s*(.+)$`)
	reSPT      = regexp.MustCompile(`^\[SPT-(S→T|L→G|Δe→∫de)\]\s+(.+)$`)
	reEEF      = regexp.MustCompile(`^\[EEF-([⊢⊨⊬⊥])\]\s+(.+)$`)
	reEvidence = regexp.MustCompile(`^\[EVIDENCE-(\d+|⊥)\]\s+(.+)$`)
	reDim      = regexp.MustCompile(`^\[DIM-([A-Za-z_-]+)\]\s+(\d+)/100\s*(?:[—:-]\s*(.+))?$`)
)

// ParseReasons walks the flat reasons array and returns a structured view.
func ParseReasons(reasons []string) ParsedReasons {
	out := ParsedReasons{}
	for _, line := range reasons {
		line = strings.TrimSpace(line)
		switch {
		case reType.MatchString(line):
			m := reType.FindStringSubmatch(line)
			out.TypeFinding = m[1]
		case reRule.MatchString(line):
			m := reRule.FindStringSubmatch(line)
			idx, _ := strconv.Atoi(m[1])
			out.Rules = append(out.Rules, RuleFinding{
				Index:   idx,
				Rule:    strings.TrimSpace(m[2]),
				Verdict: m[3],
				Reason:  strings.TrimSpace(m[4]),
			})
		case reSPT.MatchString(line):
			m := reSPT.FindStringSubmatch(line)
			out.SPTCodes = append(out.SPTCodes, m[1])
		case reEEF.MatchString(line):
			m := reEEF.FindStringSubmatch(line)
			out.EEFFlags = append(out.EEFFlags, EEFFlag{Tag: m[1], Text: m[2]})
		case reEvidence.MatchString(line):
			m := reEvidence.FindStringSubmatch(line)
			idx := -1
			if m[1] != "⊥" {
				idx, _ = strconv.Atoi(m[1])
			}
			out.Evidence = append(out.Evidence, EvidenceUse{Index: idx, Excerpt: m[2]})
		case reDim.MatchString(line):
			m := reDim.FindStringSubmatch(line)
			score, _ := strconv.Atoi(m[2])
			note := ""
			if len(m) > 3 {
				note = m[3]
			}
			out.Dimensions = append(out.Dimensions, DimensionScore{
				Name:  m[1],
				Score: score,
				Note:  note,
			})
		}
	}
	return out
}

// ToClaimAssessments derives canonical-shape ClaimAssessment records from a
// gate response. One assessment per RULE-N. Status is mapped from PASS/FAIL
// + EEF flags:
//
//	PASS + evidence present → grounded (⊢)
//	PASS without evidence    → inferred (⊨)
//	FAIL with stated reason  → extrapolated (⊬), Basis = the reason text
//	FAIL + [EEF-⊥] match     → unknown (⊥)
//
// SPT codes and confidence are response-level and attached to each claim.
func ToClaimAssessments(reasons []string, responseScore int) []schemas.ClaimAssessment {
	parsed := ParseReasons(reasons)
	if len(parsed.Rules) == 0 {
		return nil
	}

	// Map evidence by rule index when the excerpt mentions RULE-N.
	evByRule := map[int][]int{}
	for ei, ev := range parsed.Evidence {
		for _, ri := range mentionedRules(ev.Excerpt) {
			evByRule[ri] = append(evByRule[ri], ei)
		}
	}

	confidence := float64(responseScore) / 100.0

	out := make([]schemas.ClaimAssessment, 0, len(parsed.Rules))
	for _, rule := range parsed.Rules {
		var status schemas.ClaimStatus
		basis := ""
		evRefs := []string{}

		evIdxs := evByRule[rule.Index]
		for _, ei := range evIdxs {
			evRefs = append(evRefs, fmt.Sprintf("EVIDENCE-%d", parsed.Evidence[ei].Index))
		}
		if evRefs == nil {
			evRefs = []string{}
		}

		switch rule.Verdict {
		case "PASS":
			if len(evIdxs) > 0 {
				status = schemas.ClaimGrounded
			} else {
				status = schemas.ClaimInferred
			}
		case "FAIL":
			status = schemas.ClaimExtrapolated
			basis = rule.Reason

			// If any EEF-⊥ flag exists and the rule reason mentions "unknown"
			// or "cannot extract" indicators, treat as ⊥ rather than ⊬.
			if hasUnknownEEF(parsed.EEFFlags) && looksUnknown(rule.Reason) {
				status = schemas.ClaimUnknown
				basis = ""
			}
		}

		out = append(out, schemas.ClaimAssessment{
			ClaimID:       fmt.Sprintf("RULE-%d", rule.Index),
			Claim:         rule.Rule,
			Status:        status,
			Basis:         basis,
			EvidenceRefs:  evRefs,
			SPTViolations: append([]string{}, parsed.SPTCodes...),
			Confidence:    confidence,
		})
	}
	return out
}

var reRuleMention = regexp.MustCompile(`RULE-(\d+)`)

// mentionedRules returns the rule indices referenced inside an evidence excerpt.
func mentionedRules(s string) []int {
	matches := reRuleMention.FindAllStringSubmatch(s, -1)
	out := make([]int, 0, len(matches))
	for _, m := range matches {
		if n, err := strconv.Atoi(m[1]); err == nil {
			out = append(out, n)
		}
	}
	return out
}

func hasUnknownEEF(flags []EEFFlag) bool {
	for _, f := range flags {
		if f.Tag == "⊥" {
			return true
		}
	}
	return false
}

func looksUnknown(s string) bool {
	lc := strings.ToLower(s)
	for _, m := range []string{"unknown", "no data", "no record", "no evidence"} {
		if strings.Contains(lc, m) {
			return true
		}
	}
	return false
}
