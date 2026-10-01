package ui

import (
	"strconv"
	"strings"

	"github.com/yuin/goldmark/ast"
)

func renderList(l *ast.List, source []byte, width, depth int) []chatRow {
	var rows []chatRow
	counter := l.Start
	if counter == 0 {
		counter = 1
	}
	for item := l.FirstChild(); item != nil; item = item.NextSibling() {
		marker := "• "
		if l.IsOrdered() {
			marker = strconv.Itoa(counter) + ". "
			counter++
		}
		rows = append(rows, renderListItem(item, source, width, depth, marker)...)
		if !l.IsTight {
			rows = append(rows, chatRow{})
		}
	}
	return rows
}

// renderListItem renders one item's block children: the first
// paragraph/text block gets the marker as a hanging-indent prefix, a
// nested list recurses one level deeper, and anything else falls back to
// the generic block renderer, all indented under the item.
func renderListItem(item ast.Node, source []byte, width, depth int, marker string) []chatRow {
	indent := strings.Repeat("  ", depth)
	prefix := []chatSpan{{text: indent + marker, style: accent}}
	var rows []chatRow
	wroteMarker := false
	for c := item.FirstChild(); c != nil; c = c.NextSibling() {
		switch v := c.(type) {
		case *ast.List:
			rows = append(rows, renderList(v, source, width, depth+1)...)
		case *ast.Paragraph, *ast.TextBlock:
			spans := inlineSpans(c, source, base)
			if !wroteMarker {
				rows = append(rows, wrapIndented(prefix, spans, width)...)
				wroteMarker = true
			} else {
				rows = append(rows, wrapIndented([]chatSpan{{text: indent + "  "}}, spans, width)...)
			}
		default:
			rows = append(rows, renderBlock(c, source, width, depth+1)...)
		}
	}
	if !wroteMarker {
		rows = append(rows, chatRow{text: indent + marker, spans: prefix})
	}
	return rows
}
