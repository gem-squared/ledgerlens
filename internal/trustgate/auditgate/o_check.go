package auditgate

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
)

// OCheck calls the L2 post-execution gate.
// Replay semantics mirror PCheck.
func (c *Client) OCheck(
	ctx context.Context,
	req OCheckRequest,
	keys LLMKeySet,
	replayCaseID string,
) (*OCheckResponse, error) {
	body := mergeKeys(req, keys, true)

	var resp OCheckResponse
	err := c.post(ctx, "/api/audit-gate/o-check", body, &resp)
	if err == nil {
		return &resp, nil
	}

	if errors.Is(err, ErrUpstreamUnavailable) && c.Replay != nil && replayCaseID != "" {
		cached, rerr := c.Replay.LoadOCheck(replayCaseID)
		if rerr != nil {
			return nil, fmt.Errorf("auditgate: live failed (%v) and replay miss (%v)", err, rerr)
		}
		return cached, nil
	}
	return nil, err
}

// toMap round-trips through JSON to flatten a struct into map[string]any so
// downstream key-merging works regardless of the struct's original shape.
func toMap(v any) map[string]any {
	b, _ := json.Marshal(v)
	var m map[string]any
	_ = json.Unmarshal(b, &m)
	if m == nil {
		m = map[string]any{}
	}
	return m
}
