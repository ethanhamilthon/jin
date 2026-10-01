package ui

import (
	"strings"
	"testing"

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
