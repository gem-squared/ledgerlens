package memory_test

import (
	"database/sql"
	"testing"

	"github.com/gem-squared/ledgerlens/internal/trustgate/memory"
	_ "modernc.org/sqlite"
)

func TestStore_InsertCount(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()

	store, err := memory.NewStore(db)
	if err != nil {
		t.Fatalf("new store: %v", err)
	}

	cases := []memory.Record{
		{
			ID: "rid_a", GateType: "P_GATE", Verdict: "ALLOW", Score: 88, Threshold: 70,
			Input: "transfer $100", P: []string{"amount > 0"}, Evidence: []string{"chunk-1"},
			Provider: "claude", DurationMs: 7000, ReasonsJSON: `["..."]`,
		},
		{
			ID: "rid_b", GateType: "P_GATE", Verdict: "DENY", Score: 25, Threshold: 70,
			Input: "transfer bogus", P: []string{"amount > 0"}, Evidence: nil,
			Provider: "claude", DurationMs: 7000, ReasonsJSON: `["..."]`,
		},
		{
			ID: "rid_c", GateType: "O_GATE", Verdict: "SUCCESS", Score: 92, Threshold: 75,
			Input: `{"verdict":"APPROVED"}`, P: []string{"conservation"}, Evidence: []string{"chunk-1", "chunk-2"},
			Provider: "claude", DurationMs: 7000, ReasonsJSON: `["..."]`,
		},
	}
	for _, c := range cases {
		if err := store.Insert(c); err != nil {
			t.Fatalf("insert %s: %v", c.ID, err)
		}
	}

	pCount, err := store.Count("P_GATE")
	if err != nil {
		t.Fatalf("count P_GATE: %v", err)
	}
	if pCount != 2 {
		t.Errorf("P_GATE count: got %d want 2", pCount)
	}
	oCount, err := store.Count("O_GATE")
	if err != nil {
		t.Fatalf("count O_GATE: %v", err)
	}
	if oCount != 1 {
		t.Errorf("O_GATE count: got %d want 1", oCount)
	}

	// re-insert the same ID should not create a duplicate (INSERT OR REPLACE)
	if err := store.Insert(cases[0]); err != nil {
		t.Fatalf("re-insert: %v", err)
	}
	pCount, _ = store.Count("P_GATE")
	if pCount != 2 {
		t.Errorf("P_GATE count after reinsert: got %d want 2 (replace not insert)", pCount)
	}
}
