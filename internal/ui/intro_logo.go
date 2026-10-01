package ui

import "jin/internal/core"

const (
	logoEntry = "logo"
	slogan    = "a minimal and powerful terminal AI agent"
)

var logoLines = []string{
	"  ▄▄ ▄▄▄ ▄   ▄",
	"   █  █  ██▄ █",
	" ▀▄█ ▄▄▄ █ ▀██",
}

func logo(version string) chatEntry {
	return chatEntry{kind: core.UpdateInfo, tool: logoEntry, text: version}
}

// logoRows draws the logo with the version beside its last line, then the slogan.
func logoRows(version string) []chatRow {
	var rows []chatRow
	for i, line := range logoLines {
		spans := []chatSpan{{text: line, style: accent.Bold(true)}}
		if i == len(logoLines)-1 {
			spans = append(spans, chatSpan{text: "  " + version, style: dim})
		}
		rows = append(rows, chatRow{kind: core.UpdateInfo, spans: spans})
	}
	rows = append(rows, chatRow{kind: core.UpdateInfo, spans: []chatSpan{{text: " " + slogan, style: muted}}})
	return append(rows, chatRow{})
}
