package ui

import (
	east "github.com/yuin/goldmark/extension/ast"
)

type tableCellData struct {
	spans []chatSpan
	align east.Alignment
	width int
}

// minColumn is the narrowest a wrapped column may get; a table that cannot
// give every column that much is shown as records instead.
const minColumn = 8

// renderTable draws a table as a grid. A table wider than the screen wraps
// its cells; one too wide even for that lists each row as "header: value".
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
	colWidths, fits := computeColWidths(rowsData, colCount, width)
	if !fits {
		return renderRecords(rowsData, width)
	}
	var out []chatRow
	for i, r := range rowsData {
		out = append(out, renderTableRow(r, colWidths, i == 0)...)
		if i == 0 {
			out = append(out, renderTableBorder(colWidths))
		}
	}
	return out
}

// computeColWidths gives each column its natural width when the table fits.
// Otherwise narrow columns keep theirs and the wide ones share the rest; fits
// is false when a shared column would be narrower than minColumn.
func computeColWidths(rows [][]tableCellData, colCount, totalWidth int) ([]int, bool) {
	widths := make([]int, colCount)
	for _, r := range rows {
		for i, c := range r {
			widths[i] = max(widths[i], max(3, c.width))
		}
	}
	avail := totalWidth - (1 + colCount*3)
	sum := 0
	for _, w := range widths {
		sum += w
	}
	if sum <= avail {
		return widths, true
	}
	if avail < colCount*minColumn {
		return nil, false
	}
	wide, room := 0, avail
	for _, w := range widths {
		if w <= avail/colCount {
			room -= w
		} else {
			wide++
		}
	}
	share := room / max(1, wide)
	if share < minColumn {
		return nil, false
	}
	for i, w := range widths {
		if w > avail/colCount {
			widths[i] = share
		}
	}
	return widths, true
}
