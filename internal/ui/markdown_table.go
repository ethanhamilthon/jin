package ui

import (
	"strings"

	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

type tableCellData struct {
	spans []chatSpan
	align east.Alignment
	width int
}

// renderTable computes per-column max width (capped to the available terminal
// width), then renders the header, a horizontal border line, and body rows.
func renderTable(tbl *east.Table, source []byte, width int) []chatRow {
	var rowsData [][]tableCellData
	for r := tbl.FirstChild(); r != nil; r = r.NextSibling() {
		rowsData = append(rowsData, extractRowCells(r, source))
	}
	if len(rowsData) == 0 {
		return nil
	}
	colCount := 0
	for _, r := range rowsData {
		colCount = max(colCount, len(r))
	}
	if colCount == 0 {
		return nil
	}
	colWidths := computeColWidths(rowsData, colCount, width)
	var out []chatRow
	isHeader := true
	for _, r := range rowsData {
		out = append(out, renderTableRow(r, colWidths, isHeader))
		if isHeader {
			out = append(out, renderTableBorder(colWidths))
			isHeader = false
		}
	}
	return out
}

func extractRowCells(row ast.Node, source []byte) []tableCellData {
	var cells []tableCellData
	for c := row.FirstChild(); c != nil; c = c.NextSibling() {
		cell, ok := c.(*east.TableCell)
		if !ok {
			continue
		}
		spans := inlineSpans(cell, source, bodyStyle)
		cells = append(cells, tableCellData{
			spans: spans,
			align: cell.Alignment,
			width: spansWidth(spans),
		})
	}
	return cells
}

func computeColWidths(rows [][]tableCellData, colCount, totalWidth int) []int {
	widths := make([]int, colCount)
	for _, r := range rows {
		for i, c := range r {
			widths[i] = max(widths[i], max(3, c.width))
		}
	}
	overhead := 1 + colCount*3
	avail := max(colCount*3, totalWidth-overhead)
	sum := 0
	for _, w := range widths {
		sum += w
	}
	if sum > avail && sum > 0 {
		for i := range widths {
			widths[i] = max(3, widths[i]*avail/sum)
		}
	}
	return widths
}

func renderTableRow(cells []tableCellData, colWidths []int, isHeader bool) chatRow {
	bar := chatSpan{text: "│ ", style: border}
	sep := chatSpan{text: " │ ", style: border}
	end := chatSpan{text: " │", style: border}

	var spans []chatSpan
	spans = append(spans, bar)
	for i, w := range colWidths {
		if i > 0 {
			spans = append(spans, sep)
		}
		var cellSpans []chatSpan
		align := east.AlignLeft
		if i < len(cells) {
			cellSpans = cells[i].spans
			align = cells[i].align
		}
		if isHeader {
			for j := range cellSpans {
				cellSpans[j].style = tableHeaderStyle(cellSpans[j].style)
			}
		}
		spans = append(spans, padCell(cellSpans, w, align)...)
	}
	spans = append(spans, end)

	var text strings.Builder
	for _, s := range spans {
		text.WriteString(s.text)
	}
	return chatRow{text: text.String(), spans: spans}
}
