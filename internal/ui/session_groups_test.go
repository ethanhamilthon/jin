package ui

import (
	"testing"
	"time"

	"jin/internal/store"
)

func TestGroupOf(t *testing.T) {
	now := time.Date(2026, 10, 10, 15, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		updated time.Time
		running bool
		want    string
	}{
		{"running wins over age", now.Add(-48 * time.Hour), true, "Running"},
		{"just now", now.Add(-10 * time.Minute), false, "Last hour"},
		{"59 minutes", now.Add(-time.Hour + time.Second), false, "Last hour"},
		{"one hour exactly", now.Add(-time.Hour), false, "Last 6 hours"},
		{"5h59m", now.Add(-6*time.Hour + time.Second), false, "Last 6 hours"},
		{"six hours before 15:00 is still today", now.Add(-6 * time.Hour), false, "Today"},
		{"midnight is today", time.Date(2026, 10, 10, 0, 0, 0, 0, time.UTC), false, "Today"},
		{"last second of yesterday", time.Date(2026, 10, 9, 23, 59, 59, 0, time.UTC), false, "This Week"},
		{"yesterday", now.Add(-24 * time.Hour), false, "This Week"},
		{"six days", now.Add(-6 * 24 * time.Hour), false, "This Week"},
		{"seven days exactly", now.Add(-7 * 24 * time.Hour), false, "Older"},
		{"a year", now.Add(-365 * 24 * time.Hour), false, "Older"},
		{"running and a year old", now.Add(-365 * 24 * time.Hour), true, "Running"},
	}
	for _, tc := range cases {
		if got := groupOf(tc.updated, tc.running, now); got != tc.want {
			t.Errorf("%s: groupOf = %q, want %q", tc.name, got, tc.want)
		}
	}
}

func TestGroupSessionsKeepsOrderAndDropsEmptyGroups(t *testing.T) {
	now := time.Date(2026, 10, 10, 15, 0, 0, 0, time.UTC)
	rows := []store.Session{
		{ID: "run", UpdatedAt: now.Add(-3 * time.Hour)},
		{ID: "hour", UpdatedAt: now.Add(-20 * time.Minute)},
		{ID: "idle", UpdatedAt: now.Add(-2 * time.Hour)},
		{ID: "old", UpdatedAt: now.Add(-40 * 24 * time.Hour)},
		{ID: "hour2", UpdatedAt: now.Add(-5 * time.Minute)},
	}
	groups := groupSessions(rows, map[string]bool{"run": true}, now)
	var names []string
	var ids [][]string
	for _, g := range groups {
		names = append(names, g.name)
		var row []string
		for _, rec := range g.rows {
			row = append(row, rec.ID)
		}
		ids = append(ids, row)
	}
	wantNames := []string{"Running", "Last hour", "Last 6 hours", "Older"}
	if len(names) != len(wantNames) {
		t.Fatalf("groups = %v, want %v", names, wantNames)
	}
	for i := range wantNames {
		if names[i] != wantNames[i] {
			t.Fatalf("groups = %v, want %v", names, wantNames)
		}
	}
	if got := ids[1]; len(got) != 2 || got[0] != "hour" || got[1] != "hour2" {
		t.Errorf("Last hour rows = %v, want the store order [hour hour2]", got)
	}
	if got := ids[3]; len(got) != 1 || got[0] != "old" {
		t.Errorf("Older rows = %v, want [old]", got)
	}
}

func TestGroupSessionsWithNoSessions(t *testing.T) {
	if groups := groupSessions(nil, nil, time.Now()); len(groups) != 0 {
		t.Fatalf("groups = %+v, want none", groups)
	}
}
