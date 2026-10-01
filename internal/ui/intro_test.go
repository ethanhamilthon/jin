package ui

import (
	"os"
	"strings"
	"testing"

	"jin/internal/hooks"
	"jin/internal/provider"
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
	if entries[1].tool != sectionEntry || !strings.HasPrefix(entries[1].text, "Provider") {
		t.Errorf("second entry = %+v, want the Provider hint under the logo", entries[1])
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
	last := entries[len(entries)-1]
	if last.tool != sectionEntry || last.text != "Hooks\na-tone, b-style" {
		t.Errorf("last intro entry = %+v, want the Hooks section after Context", last)
	}
	if prev := entries[len(entries)-2]; !strings.HasPrefix(prev.text, "Context") {
		t.Errorf("entry before Hooks = %q, want Context", prev.text)
	}
}

func TestIntroSaysWhenNoHooksAreOn(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	a := &app{dir: t.TempDir(), registry: tools.NewRegistry(tools.NewRead())}
	entries := a.introEntries()
	if last := entries[len(entries)-1]; last.text != "Hooks\nno hooks enabled" {
		t.Errorf("last intro entry = %q", last.text)
	}
}
