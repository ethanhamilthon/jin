package ui

import "github.com/clipperhouse/displaywidth"

// moveVertical moves the cursor one visual line up (delta -1) or down (1),
// keeping its column. On the first or last line it jumps to the start or end
// of the text, like a multi-line field in a browser.
func moveVertical(input []string, cursor, width, delta int) int {
	lines, row, col := wrapInput(input, cursor, width)
	target := row + delta
	if target < 0 {
		return 0
	}
	if target >= len(lines) {
		return len(input)
	}
	start := 0
	for _, line := range lines[:target] {
		start += len(line)
		if start < len(input) && input[start] == "\n" {
			start++
		}
	}
	offset, used := 0, 0
	for _, cluster := range lines[target] {
		cells := max(1, displaywidth.String(cluster))
		if used+cells > col {
			break
		}
		used += cells
		offset++
	}
	return start + offset
}
