package ui

import (
	"slices"
	"strings"

	"github.com/clipperhouse/displaywidth"

	"jin/internal/core"
)

// promptsEntry marks the Prompts section of the intro while some prompts
// still run their commands.
const promptsEntry = "prompts"

// promptsRows draws the Prompts section: the names of the enabled prompts,
// with a purple spinner after each one whose commands are not done yet.
func promptsRows(entry chatEntry, width int) []chatRow {
	title, body, _ := strings.Cut(entry.text, "\n")
	ruleWidth := max(0, width-displaywidth.String(title)-3)
	rows := []chatRow{{spans: []chatSpan{
		{text: title + " ", style: accent.Bold(true)},
		{text: strings.Repeat("─", ruleWidth), style: border},
	}}}
	spin := base.Foreground(colorPurple)
	var spans []chatSpan
	var text strings.Builder
	used := 0
	flush := func() {
		if len(spans) > 0 {
			rows = append(rows, chatRow{kind: core.UpdateInfo, text: text.String(), spans: spans})
		}
		spans, used = nil, 0
		text.Reset()
	}
	add := func(s chatSpan) {
		spans = append(spans, s)
		shown := s.text
		if s.spin {
			shown = spinnerFrames[0]
		}
		text.WriteString(shown)
		used += displaywidth.String(shown)
	}
	names := strings.Split(body, ", ")
	for i, name := range names {
		loading := slices.Contains(entry.pending, strings.TrimPrefix(name, "#"))
		need := displaywidth.String(name) + 1
		if loading {
			need += 2
		}
		if i < len(names)-1 {
			need++
		}
		if used > 0 && used+need > max(1, width-2) {
			flush()
		}
		if used > 0 {
			add(chatSpan{text: " ", style: muted})
		}
		add(chatSpan{text: name, style: muted})
		if loading {
			add(chatSpan{text: " ", style: muted})
			add(chatSpan{text: spinnerFrames[0], style: spin, spin: true})
		}
		if i < len(names)-1 {
			add(chatSpan{text: ",", style: muted})
		}
	}
	flush()
	return append(rows, chatRow{})
}
