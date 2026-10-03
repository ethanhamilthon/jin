package ui

import (
	"strings"
	"testing"

	"github.com/clipperhouse/displaywidth"
)

const wideTable = "| Name | Description |\n|---|---|\n| read | Reads a file from disk, optionally a line range, and attaches pictures |\n"

func TestWideTableWrapsCells(t *testing.T) {
	rows := markdownRows(wideTable, 40)
	text := plain(rows)
	for _, row := range rows {
		if w := displaywidth.String(spansText(row.spans)); w > 40 {
			t.Fatalf("row wider than the screen (%d):\n%s", w, text)
		}
	}
	if !strings.Contains(text, "pictures") || strings.Contains(text, "…") {
		t.Fatalf("cell text lost:\n%s", text)
	}
	if !strings.Contains(text, "│ read") {
		t.Fatalf("not a grid:\n%s", text)
	}
}

func TestVeryWideTableBecomesRecords(t *testing.T) {
	table := "| A | B | C | D | E |\n|---|---|---|---|---|\n| one | two | three | four | five |\n| 1 | 2 | 3 | 4 | 5 |\n"
	text := plain(markdownRows(table, 30))
	for _, want := range []string{"A: one", "E: five", "C: 3"} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "│") {
		t.Fatalf("records must not draw a grid:\n%s", text)
	}
}

func TestNarrowTableKeepsItsShape(t *testing.T) {
	text := plain(markdownRows("| a | b |\n|---|---|\n| 1 | 2 |\n", 60))
	if !strings.Contains(text, "│ a   │ b   │") {
		t.Fatalf("table:\n%s", text)
	}
}
