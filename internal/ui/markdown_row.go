package ui

import (
	"strings"

	"github.com/clipperhouse/displaywidth"
)

func spansWidth(spans []chatSpan) int {
	w := 0
	for _, s := range spans {
		w += displaywidth.String(s.text)
	}
	return w
}

// prependRow rebuilds row with prefix's spans (and text) placed before it,
// keeping row.text and row.spans in lockstep the way wrapMarkdown does.
func prependRow(prefix []chatSpan, row chatRow) chatRow {
	var text strings.Builder
	for _, s := range prefix {
		text.WriteString(s.text)
	}
	text.WriteString(row.text)
	return chatRow{kind: row.kind, text: text.String(), spans: append(append([]chatSpan{}, prefix...), row.spans...)}
}

// wrapIndented wraps body to width, with prefix's own width reserved on
// every wrapped row: prefix itself (a marker like "• ") on the first row,
// matching blank padding on every continuation row (a hanging indent).
func wrapIndented(prefix []chatSpan, body []chatSpan, width int) []chatRow {
	prefixWidth := spansWidth(prefix)
	rows := wrapMarkdown(body, max(1, width-prefixWidth))
	indent := []chatSpan{{text: strings.Repeat(" ", prefixWidth)}}
	for i := range rows {
		if i == 0 {
			rows[i] = prependRow(prefix, rows[i])
		} else {
			rows[i] = prependRow(indent, rows[i])
		}
	}
	return rows
}
