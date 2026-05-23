package auditgate

import (
	"context"
	"errors"
	"fmt"
)

// PCheck calls the L1 pre-execution gate. The LLMKeys are merged into the
// request body so the upstream uses the caller's per-request key (we never
// hold long-lived LLM keys server-side).
//
// On upstream unavailability AND a Replay store wired, falls back to a
// cached response stored at artifacts/demo_cases/<replayCaseID>/p_check.json.
// Pass replayCaseID="" to skip the replay attempt.
func (c *Client) PCheck(
	ctx context.Context,
	req PCheckRequest,
	keys LLMKeySet,
	replayCaseID string,
) (*PCheckResponse, error) {
	body := mergeKeys(req, keys, false)

	var resp PCheckResponse
	err := c.post(ctx, "/api/audit-gate/p-check", body, &resp)
	if err == nil {
		return &resp, nil
	}

	// Replay fallback: only on upstream-unavailable AND case id provided.
	if errors.Is(err, ErrUpstreamUnavailable) && c.Replay != nil && replayCaseID != "" {
		cached, rerr := c.Replay.LoadPCheck(replayCaseID)
		if rerr != nil {
			return nil, fmt.Errorf("auditgate: live failed (%v) and replay miss (%v)", err, rerr)
		}
		return cached, nil
	}
	return nil, err
}

// mergeKeys builds a flat map so the request body is a single JSON object.
// (PCheckRequest + LLMKeySet flattened into one object — the gate expects
// flat fields, not nested.)
func mergeKeys(req any, keys LLMKeySet, isOCheck bool) map[string]any {
	out := toMap(req)
	if keys.Anthropic != "" {
		out["anthropic_api_key"] = keys.Anthropic
	}
	if keys.Gemini != "" {
		out["gemini_api_key"] = keys.Gemini
	}
	if keys.OpenAI != "" {
		out["openai_api_key"] = keys.OpenAI
	}
	return out
}
