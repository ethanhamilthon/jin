package ui

import (
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tools"
)

func TestBashResultKeepsLastFiveLines(t *testing.T) {
	text := resultText("bash", "1\n2\n3\n4\n5\n6\n7\n", nil)
	if !strings.HasPrefix(text, "…2 more lines\n") || !strings.HasSuffix(text, " 7") || strings.Contains(text, " 2\n") {
		t.Fatalf("text = %q", text)
	}
}

func TestEditResultShowsDiff(t *testing.T) {
	text := resultText("edit", "Replaced 1", []tools.Change{{Before: "a\nb\nc", After: "a\nB\nc"}})
	if text != "-b\n+B" {
		t.Fatalf("text = %q", text)
	}
	rows := plain(toolResultRows(text, 40))
	if !strings.Contains(rows, "│ -b") || !strings.Contains(rows, "│ +B") {
		t.Fatalf("rows:\n%s", rows)
	}
}

func TestOutputFoldShowsResultsOnlyInOutputMode(t *testing.T) {
	s := foldSession()
	s.appendEntry(chatEntry{kind: core.UpdateToolResult, tool: "bash", text: resultText("bash", "PASS ok", nil)})
	s.setFold(foldAll)
	if strings.Contains(plain(s.rows), "PASS ok") {
		t.Fatal("result shown outside output mode")
	}
	s.setFold(foldOutput)
	if !strings.Contains(plain(s.rows), "PASS ok") {
		t.Fatalf("result missing in output mode:\n%s", plain(s.rows))
	}
}

func TestSavedEditResultRebuildsDiff(t *testing.T) {
	c := provider.ToolCall{}
	c.Function.Name, c.Function.Arguments = "edit", `{"path":"f","old_string":"x","new_string":"y"}`
	if text, ok := savedResultText(c, "Replaced 1 occurrence(s) in f"); !ok || text != "-x\n+y" {
		t.Fatalf("text = %q ok=%v", text, ok)
	}
}

func TestExpandTabsDropsColors(t *testing.T) {
	if got := expandTabs("\x1b[31mred\x1b[0m\tx"); got != "red    x" {
		t.Fatalf("got %q", got)
	}
}
