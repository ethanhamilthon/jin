package ui

import (
	"strings"
	"unicode"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
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

type chatEntry struct {
	kind core.UpdateKind
	text string
	tool string
}

type chatSpan struct {
	text  string
	style tcell.Style
}

type chatRow struct {
	kind     core.UpdateKind
	text     string
	spans    []chatSpan
	fill     tcell.Style
	hasFill  bool
	fillWide bool
}

func entryRows(entry chatEntry, width int) []chatRow {
	if entry.tool == sectionEntry {
		return sectionRows(entry.text, width)
	}
	if entry.tool == logoEntry {
		return logoRows(entry.text)
	}
	switch entry.kind {
	case core.UpdateCompacted:
		return dividerRows(entry.text, width)
	case core.UpdateAssistant:
		return append(markdownRows(entry.text, width-2), chatRow{})
	case core.UpdateReasoning:
		return reasoningRows(entry.text, width)
	case core.UpdateToolCall:
		return toolCallRows(entry.tool, entry.text, width)
	case core.UpdateUser:
		return append(userRows(entry.text, width), chatRow{})
	}
	var rows []chatRow
	for _, line := range wrapChat(entry.text, width-2) {
		rows = append(rows, chatRow{kind: entry.kind, text: line})
	}
	return append(rows, chatRow{})
}
