package headless

import (
	"slices"
	"strings"
	"testing"
	"time"
)

func TestParseArgsInterleaved(t *testing.T) {
	opt, err := ParseArgs([]string{"-p", "fix", "--format", "json", "the", "bug", "-c", "--timeout", "90", "--tools", "read,bash"})
	if err != nil {
		t.Fatal(err)
	}
	if opt.Prompt != "fix the bug" || opt.Format != "json" || !opt.Continue || opt.Timeout != 90*time.Second ||
		!opt.ToolsSet || !slices.Equal(opt.Tools, []string{"read", "bash"}) {
		t.Fatalf("opt = %+v", opt)
	}
}

func TestParseArgsDoubleDash(t *testing.T) {
	opt, err := ParseArgs([]string{"-p", "--model", "m", "--", "--format", "x"})
	if err != nil || opt.Prompt != "--format x" || opt.Model != "m" {
		t.Fatalf("opt = %+v err = %v", opt, err)
	}
}

func TestParseArgsErrors(t *testing.T) {
	for name, args := range map[string][]string{
		"format":   {"-p", "--format", "xml", "hi"},
		"both":     {"-p", "-c", "--session", "abc", "hi"},
		"timeout":  {"-p", "--timeout", "soon", "hi"},
		"zero":     {"-p", "--timeout", "0", "hi"},
		"empty":    {"-p", "--tools", "read,,bash", "hi"},
		"unknown":  {"-p", "--nope", "hi"},
		"no value": {"-p", "hi", "--model"},
	} {
		if _, err := ParseArgs(args); err == nil {
			t.Errorf("%s: no error", name)
		}
	}
}

func TestTimeoutUnits(t *testing.T) {
	for in, want := range map[string]time.Duration{"2s": 2 * time.Second, "10m": 10 * time.Minute, "5": 5 * time.Second, "1.5": 1500 * time.Millisecond} {
		if got, err := parseTimeout(in); err != nil || got != want {
			t.Errorf("%q = %v, %v", in, got, err)
		}
	}
}

func TestHandles(t *testing.T) {
	for args, want := range map[string]bool{"-p hi": true, "hi -p": true, "models": true, "refresh-models --efforts": true, "": false, "--version": false, "hi": false} {
		if got := Handles(strings.Fields(args)); got != want {
			t.Errorf("%q = %v", args, got)
		}
	}
}

func TestBuildPrompt(t *testing.T) {
	got, err := BuildPrompt(t.Context(), "summarize", strings.NewReader("data\n"), true)
	if err != nil || got != "summarize\n\n<stdin>\ndata\n</stdin>" {
		t.Fatalf("got %q, %v", got, err)
	}
	if got, _ := BuildPrompt(t.Context(), "", strings.NewReader(" hi \n"), true); got != "hi" {
		t.Fatalf("stdin prompt = %q", got)
	}
	if _, err := BuildPrompt(t.Context(), "", nil, false); err == nil {
		t.Fatal("empty terminal prompt accepted")
	}
	if _, err := BuildPrompt(t.Context(), "x", strings.NewReader(strings.Repeat("a", maxStdin+1)), true); err == nil {
		t.Fatal("oversized stdin accepted")
	}
	if got, _ := BuildPrompt(t.Context(), "only", strings.NewReader(""), true); got != "only" {
		t.Fatalf("empty pipe = %q", got)
	}
}

func TestToolNames(t *testing.T) {
	got, err := toolNames(nil, Options{})
	if err != nil || slices.Contains(got, "ask_user") || !slices.Contains(got, "todo") {
		t.Fatalf("default = %v, %v", got, err)
	}
	got, _ = toolNames([]string{"bash"}, Options{ToolsSet: true, Tools: []string{"bash", "read"}})
	if !slices.Equal(got, []string{"read"}) {
		t.Fatalf("allowlist cannot enable a disabled tool: %v", got)
	}
	got, _ = toolNames(nil, Options{Exclude: []string{"bash", "todo"}})
	if slices.Contains(got, "bash") || slices.Contains(got, "todo") {
		t.Fatalf("exclude = %v", got)
	}
	if got, _ = toolNames(nil, Options{NoTools: true}); len(got) != 0 {
		t.Fatalf("no-tools = %v", got)
	}
	if _, err := toolNames(nil, Options{Tools: []string{"nope"}, ToolsSet: true}); err == nil {
		t.Fatal("unknown tool accepted")
	}
}
