package ui

import "github.com/gdamore/tcell/v3"

func paintSelection(screen tcell.Screen, sel textSelection, firstRow, chatHeight int) {
	if !sel.active || chatHeight <= 0 {
		return
	}
	startRow, startCol := sel.startRow, sel.startCol
	endRow, endCol := sel.endRow, sel.endCol
	if startRow > endRow || (startRow == endRow && startCol > endCol) {
		startRow, endRow = endRow, startRow
		startCol, endCol = endCol, startCol
	}
	if startRow == endRow && startCol == endCol {
		return
	}
	w, _ := screen.Size()
	for row := max(startRow, firstRow); row <= min(endRow, firstRow+chatHeight-1); row++ {
		y := row - firstRow
		left, right := 1, w
		if row == startRow {
			left = startCol
		}
		if row == endRow {
			right = endCol
		}
		if row != startRow || row != endRow {
			// The buffer has no row metadata; stop at the last visible glyph.
			// On a body row the first column is an unrendered margin.
			if glyph, _, _ := screen.Get(1, y); glyph == " " {
				left = max(left, 2)
			}
			if row != endRow {
				right = lastVisibleCell(screen, y, w)
			}
		}
		for x := max(1, left-1); x < min(w, right); {
			glyph, style, width := screen.Get(x, y)
			width = max(1, width)
			if x+width > left && glyph != "" {
				screen.Put(x, y, glyph, style.Reverse(true))
			}
			x += width
		}
	}
}

func lastVisibleCell(screen tcell.Screen, y, width int) int {
	end := 1
	for x := 1; x < width; {
		glyph, _, cells := screen.Get(x, y)
		cells = max(1, cells)
		if glyph != " " && glyph != "" {
			end = x + cells
		}
		x += cells
	}
	return end
}
