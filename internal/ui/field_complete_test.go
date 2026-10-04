package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

// TestProjectDirectoryFieldCompletesDirectories opens the add-directory field
// of the projects panel and checks that it offers directories only, and that
// Tab puts the highlighted one back in the field.
func TestProjectDirectoryFieldCompletesDirectories(t *testing.T) {
	a := startingApp(t)
	a.dir = t.TempDir()
	for _, dir := range []string{"docs", "drafts"} {
		if err := os.Mkdir(filepath.Join(a.dir, dir), 0o700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(a.dir, "notes.txt"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	a.openProjects()
	a.sel.actions['a']("")
	if a.sel == nil || a.sel.complete == nil {
		t.Fatalf("a must open a completing directory field, got %+v", a.sel)
	}
	typeRune(a, "d")
	if got := candidateLabels(a.sel); strings.Join(got, ",") != "docs/,drafts/" {
		t.Fatalf("candidates = %v, want the two directories", got)
	}
	a.selectorKey(tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone))
	if got := strings.Join(a.sel.query, ""); got != "docs/" {
		t.Fatalf("Tab left query %q, want the first directory", got)
	}
	if got := candidateLabels(a.sel); len(got) != 0 {
		t.Fatalf("candidates inside an empty directory = %v", got)
	}
}

// TestFieldWithoutCompletionKeepsTyping checks that a plain field is not
// turned into a completion field.
func TestFieldWithoutCompletionKeepsTyping(t *testing.T) {
	a, _ := layoutApp(t)
	a.openField("Name", "", false, func(string) error { return nil })
	typeRune(a, "x")
	a.selectorKey(tcell.NewEventKey(tcell.KeyTab, "", tcell.ModNone))
	if got := strings.Join(a.sel.query, ""); got != "x" {
		t.Fatalf("query = %q, want the typed text only", got)
	}
}

func candidateLabels(sel *selector) []string {
	labels := make([]string, len(sel.cands))
	for i, opt := range sel.cands {
		labels[i] = opt.label
	}
	return labels
}
