package ui

import (
	"os"
	"strings"
	"testing"

	"jin/internal/hooks"
	"jin/internal/provider"
	"jin/internal/sysprompt"
	"jin/internal/tools"
)

func TestLogoRowsShowVersionAndSlogan(t *testing.T) {
	rows := logoRows("v0.1")
	var lines []string
	for _, row := range rows {
		var b strings.Builder
		for _, span := range row.spans {
			b.WriteString(span.text)
		}
		lines = append(lines, b.String())
	}
	if len(lines) != len(logoLines)+1 {
		t.Fatalf("rows = %d, want %d", len(lines), len(logoLines)+1)
	}
	if !strings.HasSuffix(lines[len(logoLines)-1], "v0.1") {
		t.Errorf("version missing on the last logo line: %q", lines[len(logoLines)-1])
	}
	if !strings.HasSuffix(lines[0], slogan) {
		t.Errorf("slogan missing on the first logo line: %q", lines[0])
	}
}

func TestIntroStartsWithLogo(t *testing.T) {
	a := &app{version: "v0.1", registry: tools.NewRegistry(tools.NewRead())}
	entries := a.introEntries()
	if entries[0].tool != logoEntry || entries[0].text != "v0.1" {
		t.Fatalf("first entry = %+v, want the logo", entries[0])
	}
	if entries[1].tool != sectionEntry || !strings.HasPrefix(entries[1].text, "Tools") {
		t.Errorf("second entry = %+v, want Tools under the logo", entries[1])
	}
}

func TestIntroListsActiveHooksAfterContext(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for name, body := range map[string]string{"b-style": "Be brief.", "a-tone": "Be kind.", "off": "Hidden.", "empty": ""} {
		path, err := hooks.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	a := &app{dir: t.TempDir(), registry: tools.NewRegistry(tools.NewRead())}
	a.cfg.Provider = provider.Config{BaseURL: "http://x", APIKey: "k"}
	a.cfg.HooksDisabled = []string{"off"}
	entries := a.introEntries()
	hooks := entries[len(entries)-2]
	if hooks.tool != sectionEntry || hooks.text != "Hooks\na-tone, b-style" {
		t.Errorf("intro entry = %+v, want the Hooks section after Context", hooks)
	}
	if prev := entries[len(entries)-3]; !strings.HasPrefix(prev.text, "Context") {
		t.Errorf("entry before Hooks = %q, want Context", prev.text)
	}
	if last := entries[len(entries)-1]; last.tool != promptsEntry || last.text != "Prompts\n#plan, #review" {
		t.Errorf("last intro entry = %+v, want the Prompts section", last)
	}
}

func TestIntroSaysWhenNoHooksAreOn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a := &app{dir: t.TempDir(), registry: tools.NewRegistry(tools.NewRead())}
	entries := a.introEntries()
	if hooks := entries[len(entries)-2]; hooks.text != "Hooks\nno hooks enabled" {
		t.Errorf("hooks intro entry = %q", hooks.text)
	}
}

func TestIntroListsOnlyEnabledPrompts(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a := &app{dir: t.TempDir(), registry: tools.NewRegistry(tools.NewRead())}
	a.cfg.PromptsDisabled = []string{"review"}
	entries := a.introEntries()
	if last := entries[len(entries)-1]; last.text != "Prompts\n#plan" {
		t.Errorf("prompts intro entry = %q", last.text)
	}
	a.cfg.PromptsDisabled = []string{"plan", "review"}
	if last := a.introEntries(); last[len(last)-1].text != "Prompts\nno prompts enabled" {
		t.Errorf("prompts intro entry = %q", last[len(last)-1].text)
	}
}

func TestIntroMarksACustomSystemPrompt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a := &app{dir: t.TempDir(), registry: tools.NewRegistry(tools.NewRead())}
	has := func() (string, bool) {
		for _, e := range a.introEntries() {
			if strings.HasPrefix(e.text, "System prompt\n") {
				return e.text, true
			}
		}
		return "", false
	}
	if text, ok := has(); ok {
		t.Fatalf("no file, no section, got %q", text)
	}
	path, err := sysprompt.Create()
	if err != nil {
		t.Fatal(err)
	}
	text, ok := has()
	if !ok || !strings.Contains(text, "custom (") || !strings.Contains(text, "system-prompt.md") {
		t.Errorf("section = %q, %v", text, ok)
	}
	os.Remove(path)
	if _, ok := has(); ok {
		t.Error("deleting the file must reset the section")
	}
}
