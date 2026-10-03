package ui

import (
	"github.com/yuin/goldmark/ast"
	east "github.com/yuin/goldmark/extension/ast"
)

func extractRowCells(row ast.Node, source []byte) []tableCellData {
	var cells []tableCellData
	for c := row.FirstChild(); c != nil; c = c.NextSibling() {
		cell, ok := c.(*east.TableCell)
		if !ok {
			continue
		}
		spans := inlineSpans(cell, source, bodyStyle)
		cells = append(cells, tableCellData{spans: spans, align: cell.Alignment, width: spansWidth(spans)})
	}
	return cells
}
