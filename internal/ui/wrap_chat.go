package ui

import (
	"strings"
	"unicode"

	"github.com/clipperhouse/displaywidth"
)

// wrapChat splits on grapheme boundaries, preserving explicit line breaks.
func wrapChat(text string, width int) []string {
	if width < 1 {
		return []string{""}
	}
	var rows []string
	for _, line := range strings.Split(text, "\n") {
		line = strings.Map(func(r rune) rune {
			if unicode.IsControl(r) {
				return ' '
			}
			return r
		}, line)
		var row strings.Builder
		x := 0
		graphemes := displaywidth.StringGraphemes(line)
		for graphemes.Next() {
			cluster, cells := graphemes.Value(), graphemes.Width()
			if cells < 1 {
				continue
			}
			if cells > width {
				cluster, cells = "?", 1
			}
			if x+cells > width {
				rows = append(rows, row.String())
				row.Reset()
				x = 0
			}
			row.WriteString(cluster)
			x += cells
		}
		rows = append(rows, row.String())
	}
	return rows
}
