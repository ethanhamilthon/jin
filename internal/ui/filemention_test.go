package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
)

func fileApp(t *testing.T) (*app, string) {
	t.Helper()
	a, _ := layoutApp(t)
	db, _ := openFoldDB(t)
	a.store, a.active.store, a.active.path = db, db, "p"
	dir := t.TempDir()
	for _, name := range []string{"notes.md", "my file.md", "main.go"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.Mkdir(filepath.Join(dir, "src"), 0o755); err != nil {
		t.Fatal(err)
	}
	a.dir = dir
	return a, dir
}

func TestAtOpensFileListAndTabCompletes(t *testing.T) {
	a, _ := fileApp(t)
	typeText(a, "look at @no")
	if a.file == nil {
		t.Fatal("@no should open the file list")
	}
	press(a, tcell.KeyTab)
	if got := strings.Join(a.active.input, ""); got != "look at @notes.md " {
		t.Fatalf("draft = %q", got)
	}
}

func TestAtQuotesNamesWithSpaces(t *testing.T) {
	a, _ := fileApp(t)
	typeText(a, "@my")
	press(a, tcell.KeyTab)
	if got := strings.Join(a.active.input, ""); got != `@"my file.md" ` {
		t.Fatalf("draft = %q", got)
	}
}

func TestAtQuotedTokenStillCompletesWhileOpen(t *testing.T) {
	a, _ := fileApp(t)
	typeText(a, `@"my`)
	if a.file == nil {
		t.Fatal("an open quote should keep the list open")
	}
}

func TestAtDirectoryKeepsTheListGoingInside(t *testing.T) {
	a, dir := fileApp(t)
	if err := os.WriteFile(filepath.Join(dir, "src", "app.go"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	typeText(a, "@sr")
	press(a, tcell.KeyTab)
	if got := strings.Join(a.active.input, ""); got != "@src/" {
		t.Fatalf("draft = %q", got)
	}
	a.refreshFile()
	if a.file == nil || a.file.sel.options[0].value != "src/app.go" {
		t.Fatalf("list should show the directory contents: %+v", a.file)
	}
}

func TestAtMissingPathStaysPlainText(t *testing.T) {
	a, _ := fileApp(t)
	typeText(a, "mail me@example.com or @nothing-here")
	if a.file != nil {
		t.Fatal("no file matches, no list expected")
	}
}

func TestAtEscKeepsTokenAsText(t *testing.T) {
	a, _ := fileApp(t)
	typeText(a, "@no")
	press(a, tcell.KeyEscape)
	if a.file != nil || strings.Join(a.active.input, "") != "@no" {
		t.Fatal("Esc should close the list and keep the text")
	}
}

func TestSendAddsAttachedFilesBlockAndShowsDraftAsTyped(t *testing.T) {
	a, dir := fileApp(t)
	a.active.input = clusters("read @notes.md and @missing.md")
	a.sendDraft(strings.Join(a.active.input, ""))
	if len(a.active.pending) != 1 {
		t.Fatal("expected one queued request")
	}
	prompt := a.active.pending[0].Prompt
	notes := filepath.Join(dir, "notes.md")
	if !strings.Contains(prompt, "read "+notes+" and @missing.md") {
		t.Fatalf("prompt = %q", prompt)
	}
	if !strings.Contains(prompt, `<file path="`+notes+`"/>`) {
		t.Fatalf("no attached-files block: %q", prompt)
	}
	if last := a.active.history[len(a.active.history)-1]; last.text != "read @notes.md and @missing.md" {
		t.Fatalf("chat shows %q", last.text)
	}
}

func TestSendWithoutFilesHasNoBlock(t *testing.T) {
	a, _ := fileApp(t)
	a.sendDraft("hello")
	if strings.Contains(a.active.pending[0].Prompt, "attached-files") {
		t.Fatal("no block expected")
	}
}
