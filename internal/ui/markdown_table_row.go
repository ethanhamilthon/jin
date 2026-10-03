package ui

import (
	"strconv"
	"strings"

	east "github.com/yuin/goldmark/extension/ast"
)

// renderTableRow draws one table row. A cell longer than its column wraps,
// so the row may take several screen rows.
func renderTableRow(cells []tableCellData, colWidths []int, isHeader bool) []chatRow {
	lines := make([][]chatRow, len(colWidths))
	height := 1
	for i, w := range colWidths {
		var cellSpans []chatSpan
		if i < len(cells) {
			cellSpans = cells[i].spans
		}
		if isHeader {
			for j := range cellSpans {
				cellSpans[j].style = tableHeaderStyle(cellSpans[j].style)
			}
		}
		if spansWidth(cellSpans) > w {
			lines[i] = wrapMarkdown(cellSpans, w)
		} else {
			lines[i] = []chatRow{{spans: cellSpans}}
		}
		height = max(height, len(lines[i]))
	}
	out := make([]chatRow, height)
	for line := range height {
		spans := []chatSpan{{text: "│ ", style: border}}
		for i, w := range colWidths {
			if i > 0 {
				spans = append(spans, chatSpan{text: " │ ", style: border})
			}
			align := east.AlignLeft
			if i < len(cells) {
				align = cells[i].align
			}
			var part []chatSpan
			if line < len(lines[i]) {
				part = lines[i][line].spans
			}
			spans = append(spans, padCell(part, w, align)...)
		}
		spans = append(spans, chatSpan{text: " │", style: border})
		out[line] = chatRow{text: spansText(spans), spans: spans}
	}
	return out
}

// renderRecords lists a table too wide for the screen as one block per row,
// each cell on its own line after its header.
func renderRecords(rows [][]tableCellData, width int) []chatRow {
	header := rows[0]
	var out []chatRow
	for n, row := range rows[1:] {
		if n > 0 {
			out = append(out, chatRow{spans: []chatSpan{{text: strings.Repeat("─", max(1, min(width, 20))), style: border}}})
		}
		for i, cell := range row {
			label := "column " + strconv.Itoa(i+1)
			if i < len(header) {
				label = spansText(header[i].spans)
			}
			prefix := []chatSpan{{text: label + ": ", style: tableHeaderStyle(bodyStyle)}}
			if spansWidth(prefix) > width/3 {
				out = append(out, wrapMarkdown(prefix, width)...)
				out = append(out, wrapIndented([]chatSpan{{text: "  "}}, cell.spans, width)...)
				continue
			}
			out = append(out, wrapIndented(prefix, cell.spans, width)...)
		}
	}
	return out
}

func spansText(spans []chatSpan) string {
	var text strings.Builder
	for _, s := range spans {
		text.WriteString(s.text)
	}
	return text.String()
}
