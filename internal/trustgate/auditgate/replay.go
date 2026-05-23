package auditgate

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ReplayStore is the file-backed cache used for offline/demo-stability fallback.
// Layout:
//
//	<root>/<case_id>/p_check.json   ← prior PCheckResponse
//	<root>/<case_id>/o_check.json   ← prior OCheckResponse
//
// Saved by SavePCheck/SaveOCheck (typically after a live success at demo-prep
// time). Loaded by LoadPCheck/LoadOCheck when the upstream is unreachable.
type ReplayStore struct {
	Root string
}

func NewReplayStore(root string) (*ReplayStore, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("replay store: mkdir %s: %w", root, err)
	}
	return &ReplayStore{Root: root}, nil
}

func (s *ReplayStore) caseDir(caseID string) string {
	return filepath.Join(s.Root, caseID)
}

func (s *ReplayStore) SavePCheck(caseID string, resp *PCheckResponse) error {
	return s.save(caseID, "p_check.json", resp)
}

func (s *ReplayStore) SaveOCheck(caseID string, resp *OCheckResponse) error {
	return s.save(caseID, "o_check.json", resp)
}

func (s *ReplayStore) save(caseID, name string, v any) error {
	dir := s.caseDir(caseID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("replay store: mkdir %s: %w", dir, err)
	}
	path := filepath.Join(dir, name)
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("replay store: marshal: %w", err)
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		return fmt.Errorf("replay store: write %s: %w", path, err)
	}
	return nil
}

func (s *ReplayStore) LoadPCheck(caseID string) (*PCheckResponse, error) {
	path := filepath.Join(s.caseDir(caseID), "p_check.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("replay store: read %s: %w", path, err)
	}
	var out PCheckResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("replay store: unmarshal: %w", err)
	}
	return &out, nil
}

func (s *ReplayStore) LoadOCheck(caseID string) (*OCheckResponse, error) {
	path := filepath.Join(s.caseDir(caseID), "o_check.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("replay store: read %s: %w", path, err)
	}
	var out OCheckResponse
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("replay store: unmarshal: %w", err)
	}
	return &out, nil
}
