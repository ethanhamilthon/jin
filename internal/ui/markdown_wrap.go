package ui

import (
	"strings"
	"unicode"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

// wrapMarkdown keeps the displayed text and its styled fragments in lockstep.
// A line break span starts a new row.
func wrapMarkdown(spans []chatSpan, width int) []chatRow {
	if width < 1 {
		return []chatRow{{kind: core.UpdateAssistant}}
	}
	var rows []chatRow
	var rowText, runText strings.Builder
	var fragments []chatSpan
	var activeStyle tcell.Style
	cells, active := 0, false
	flushRun := func() {
		if runText.Len() > 0 {
			fragments = append(fragments, chatSpan{text: runText.String(), style: activeStyle})
			runText.Reset()
		}
	}
	flushRow := func() {
		flushRun()
		rows = append(rows, chatRow{kind: core.UpdateAssistant, text: rowText.String(), spans: fragments})
		rowText.Reset()
		fragments = nil
		cells, active = 0, false
	}
	add := func(cluster string, size int, style tcell.Style) {
		if cells+size > width {
			flushRow()
		}
		if !active || style != activeStyle {
			flushRun()
			activeStyle, active = style, true
		}
		rowText.WriteString(cluster)
		runText.WriteString(cluster)
		cells += size
	}
	for _, span := range spans {
		if span.lineBreak {
			flushRow()
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
				for n := 4 - cells%4; n > 0; n-- {
					add(" ", 1, span.style)
				}
				continue
			}
			if size < 1 {
				continue
			}
			if size > width {
				cluster, size = "?", 1
			}
			add(cluster, size, span.style)
		}
	}
	flushRow()
	return rows
}
