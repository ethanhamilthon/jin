package store

import (
	"fmt"
	"testing"
	"time"
)

func TestSearchSessionsLimitAndOrder(t *testing.T) {
	db := newTestDB(t)
	dir := "/test/dir"
	baseTime := time.Now().Unix()

	// Insert 25 matching sessions with increasing updated_at
	for i := 1; i <= 25; i++ {
		id := fmt.Sprintf("s%02d", i)
		title := fmt.Sprintf("Session %d keyword", i)
		if err := db.Touch(id, dir, "model", "high", title); err != nil {
			t.Fatalf("Touch %s: %v", id, err)
		}
		_, err := db.sql.Exec(`UPDATE sessions SET updated_at = ? WHERE id = ?`, baseTime+int64(i), id)
		if err != nil {
			t.Fatalf("Exec updated_at %s: %v", id, err)
		}
	}

	results, err := db.SearchSessions(dir, false, []string{"keyword"})
	if err != nil {
		t.Fatalf("SearchSessions: %v", err)
	}

	if len(results) != 20 {
		t.Fatalf("expected exactly 20 results (limit cap), got %d", len(results))
	}

	// Verify exact newest-first order: s25 down to s06
	for i := 0; i < 20; i++ {
		expectedID := fmt.Sprintf("s%02d", 25-i)
		if results[i].ID != expectedID {
			t.Errorf("result[%d] = %s, want %s", i, results[i].ID, expectedID)
		}
	}
}
