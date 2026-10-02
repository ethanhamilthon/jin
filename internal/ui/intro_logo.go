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

// logoRows draws the logo with the slogan beside its first line and the version beside its last.
func logoRows(version string) []chatRow {
	var rows []chatRow
	for i, line := range logoLines {
		spans := []chatSpan{{text: line, style: accent.Bold(true), shimmer: true}}
		if i == 0 {
			spans = append(spans, chatSpan{text: "  " + slogan, style: accent})
		}
		if i == len(logoLines)-1 {
			spans = append(spans, chatSpan{text: "  " + version, style: dim})
		}
		rows = append(rows, chatRow{kind: core.UpdateInfo, spans: spans})
	}
	return append(rows, chatRow{})
}
