package ui

import (
	"strings"

	"jin/internal/core"
)

// asyncRows draws a task result: a purple headline, then the first lines of
// the output, muted.
func asyncRows(text string, width int) []chatRow {
	headline, body, _ := strings.Cut(text, "\n")
	rows := []chatRow{{spans: []chatSpan{{text: "◐ " + headline, style: base.Foreground(colorPurple).Bold(true)}}}}
	for _, line := range strings.Split(body, "\n") {
		for _, wrapped := range wrapChat(line, width-2) {
			rows = append(rows, chatRow{text: wrapped, kind: core.UpdateInfo})
		}
	}
	return append(rows, chatRow{})
}
