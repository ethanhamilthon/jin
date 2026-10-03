package ui

import (
	"strings"
	"unicode"

	"github.com/clipperhouse/displaywidth"
)

// wrapChat wraps text by words, preserving explicit line breaks. A word
// wider than the row is split on grapheme boundaries.
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
		var clusters []string
		var widths []int
		var spaces []bool
		graphemes := displaywidth.StringGraphemes(line)
		for graphemes.Next() {
			cluster, cells := graphemes.Value(), graphemes.Width()
			if cells < 1 {
				continue
			}
			if cells > width {
				cluster, cells = "?", 1
			}
			clusters = append(clusters, cluster)
			widths = append(widths, cells)
			spaces = append(spaces, isSpace(cluster))
		}
		for _, b := range rowBounds(wordBreaks(widths, spaces, width), len(clusters), spaces, true) {
			rows = append(rows, strings.Join(clusters[b[0]:b[1]], ""))
		}
	}
	return rows
}
