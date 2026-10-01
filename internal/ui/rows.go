package ui

import (
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

const sectionEntry = "section"

func rowStyle(kind core.UpdateKind) tcell.Style {
	switch kind {
	case core.UpdateError:
		return errorStyle
	case core.UpdateReasoning, core.UpdateInfo:
		return muted
	default:
		return base
	}
}

func userRows(text string, width int) []chatRow {
	var rows []chatRow
	for _, line := range wrapChat(text, width-2) {
		rows = append(rows, chatRow{kind: core.UpdateUser, text: line, fill: userStyle, hasFill: true, fillWide: true})
	}
	return rows
}

func toolCallRows(tool, argument string, width int) []chatRow {
	main, detail := splitTrailingDetail(argument)
	spans := []chatSpan{
		{text: tool, style: toolStyle.Foreground(toolAccent(tool)).Bold(true)},
		{text: "  " + main, style: toolStyle.Foreground(colorArgument)},
	}
	if detail != "" {
		spans = append(spans, chatSpan{text: " " + detail, style: toolStyle.Foreground(colorDetail)})
	}
	rows := wrapMarkdown(spans, width-2)
	for i := range rows {
		rows[i].kind = core.UpdateToolCall
		rows[i].fill, rows[i].hasFill, rows[i].fillWide = toolStyle, true, true
	}
	return rows
}

// splitTrailingDetail peels a trailing " (...)" such as bash's timeout off a
// summary so it can render dimmer than the argument itself.
func splitTrailingDetail(s string) (main, detail string) {
	if strings.HasSuffix(s, ")") {
		if i := strings.LastIndex(s, " ("); i >= 0 {
			return s[:i], s[i+1:]
		}
	}
	return s, ""
}

// reasoningRows shows thinking as one clipped line ending in "..." when cut.
func reasoningRows(text string, width int) []chatRow {
	line, _, cut := strings.Cut(strings.TrimSpace(text), "\n")
	style := dim.Italic(true)
	avail := max(1, width-2)
	if cut || displaywidth.String(line) > avail {
		line = truncateWidth(line, max(1, avail-3)) + "..."
	}
	return []chatRow{{kind: core.UpdateReasoning, text: line, spans: []chatSpan{{text: line, style: style}}}, {}}
}

// sectionRows renders "Title ────────" followed by the wrapped body.
func sectionRows(text string, width int) []chatRow {
	title, body, _ := strings.Cut(text, "\n")
	ruleWidth := max(0, width-displaywidth.String(title)-3)
	rows := []chatRow{{spans: []chatSpan{
		{text: title + " ", style: accent.Bold(true)},
		{text: strings.Repeat("─", ruleWidth), style: border},
	}}}
	for _, line := range strings.Split(body, "\n") {
		for _, wrapped := range wrapChat(line, width-2) {
			rows = append(rows, chatRow{text: wrapped, kind: core.UpdateInfo})
		}
	}
	return append(rows, chatRow{})
}
