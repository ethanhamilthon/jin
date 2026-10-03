package export

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func exportDB(t *testing.T) *store.DB {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	db, err := store.Open()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	_ = db.Touch("abcdef", "/repo", "gpt", "", "Fix tests")
	call := provider.ToolCall{ID: "1"}
	call.Function.Name, call.Function.Arguments = "bash", `{"command":"ls"}`
	for _, m := range []provider.Message{
		{Role: "user", Content: "run ls"},
		{Role: "assistant", ToolCalls: []provider.ToolCall{call}},
		{Role: "tool", ToolCallID: "1", Content: "a.go"},
		{Role: "assistant", Content: "done"},
	} {
		_ = db.AppendMessage("abcdef", m)
	}
	return db
}

func TestExportMarkdownByPrefix(t *testing.T) {
	db := exportDB(t)
	var out, errOut bytes.Buffer
	if code := Main([]string{"abc"}, db, &out, &errOut); code != 0 {
		t.Fatalf("code %d: %s", code, errOut.String())
	}
	for _, want := range []string{"# Fix tests", "## User\n\nrun ls", "**bash**", "a.go", "## Assistant\n\ndone"} {
		if !strings.Contains(out.String(), want) {
			t.Errorf("missing %q in:\n%s", want, out.String())
		}
	}
}

func TestExportJSON(t *testing.T) {
	db := exportDB(t)
	var out, errOut bytes.Buffer
	if code := Main([]string{"abcdef", "--json"}, db, &out, &errOut); code != 0 {
		t.Fatalf("code %d: %s", code, errOut.String())
	}
	var got jsonExport
	if err := json.Unmarshal(out.Bytes(), &got); err != nil || got.Title != "Fix tests" || len(got.Messages) != 4 {
		t.Fatalf("got %+v err %v", got, err)
	}
}

func TestExportErrors(t *testing.T) {
	db := exportDB(t)
	var out, errOut bytes.Buffer
	if code := Main(nil, db, &out, &errOut); code != 2 {
		t.Errorf("no id: code %d", code)
	}
	if code := Main([]string{"zzz"}, db, &out, &errOut); code != 1 || !strings.Contains(errOut.String(), "no session") {
		t.Errorf("unknown id: code %d %s", code, errOut.String())
	}
}
