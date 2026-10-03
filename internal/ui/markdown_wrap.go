package ui

import (
	"strings"
	"unicode"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

type styledCluster struct {
	text  string
	cells int
	style tcell.Style
}

// wrapMarkdown keeps the displayed text and its styled fragments in lockstep.
// Rows break between words; a line break span starts a new row.
func wrapMarkdown(spans []chatSpan, width int) []chatRow {
	if width < 1 {
		return []chatRow{{kind: core.UpdateAssistant}}
	}
	var rows []chatRow
	var line []styledCluster
	column := 0
	flushLine := func() {
		rows = append(rows, layoutLine(line, width)...)
		line, column = nil, 0
	}
	for _, span := range spans {
		if span.lineBreak {
			flushLine()
			continue
		}
		// A terminal should never receive control bytes, including escape codes.
		clean := strings.Map(func(r rune) rune {
			if unicode.IsControl(r) && r != '\t' {
				return ' '
			}
			return r
		}, span.text)
		graphemes := displaywidth.StringGraphemes(clean)
		for graphemes.Next() {
			cluster, size := graphemes.Value(), graphemes.Width()
			if cluster == "\t" {
				for n := 4 - column%4; n > 0; n-- {
					line = append(line, styledCluster{" ", 1, span.style})
					column++
				}
				continue
			}
			if size < 1 {
				continue
			}
			if size > width {
				cluster, size = "?", 1
			}
			line = append(line, styledCluster{cluster, size, span.style})
			column += size
		}
	}
	flushLine()
	return rows
}

// layoutLine word-wraps one source line into rows of styled fragments.
func layoutLine(line []styledCluster, width int) []chatRow {
	widths, spaces := make([]int, len(line)), make([]bool, len(line))
	for i, c := range line {
		widths[i], spaces[i] = c.cells, isSpace(c.text)
	}
	var rows []chatRow
	for _, b := range rowBounds(wordBreaks(widths, spaces, width), len(line), spaces, true) {
		var text strings.Builder
		var fragments []chatSpan
		for _, c := range line[b[0]:b[1]] {
			text.WriteString(c.text)
			if n := len(fragments); n > 0 && fragments[n-1].style == c.style {
				fragments[n-1].text += c.text
				continue
			}
			fragments = append(fragments, chatSpan{text: c.text, style: c.style})
		}
		rows = append(rows, chatRow{kind: core.UpdateAssistant, text: text.String(), spans: fragments})
	}
	return rows
}
