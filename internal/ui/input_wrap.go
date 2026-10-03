package ui

import "github.com/clipperhouse/displaywidth"

// wrapInput wraps input to width, honoring literal "\n" clusters (inserted
// by Shift+Enter) as forced line breaks in addition to width wrapping.
func wrapInput(input []string, cursor, width int) (lines [][]string, cursorRow, cursorCol int) {
	if width < 1 {
		width = 1
	}
	segStart, found := 0, false
	for i := 0; i <= len(input); i++ {
		if i != len(input) && input[i] != "\n" {
			continue
		}
		seg := input[segStart:i]
		local := max(0, min(len(seg), cursor-segStart))
		segLines, segRow, segCol := wrapSegment(seg, local, width)
		if !found && cursor >= segStart && cursor <= i {
			cursorRow, cursorCol, found = len(lines)+segRow, segCol, true
		}
		lines = append(lines, segLines...)
		segStart = i + 1
	}
	if len(lines) == 0 {
		lines = [][]string{{}}
	}
	return
}

// wrapSegment wraps one newline-free run of clusters to width, between
// words when it can. Every cluster stays, so the cursor maps one to one.
func wrapSegment(input []string, cursor, width int) (lines [][]string, cursorRow, cursorCol int) {
	widths, spaces := make([]int, len(input)), make([]bool, len(input))
	for i, cluster := range input {
		widths[i], spaces[i] = max(1, displaywidth.String(cluster)), isSpace(cluster)
	}
	breaks := wordBreaks(widths, spaces, width)
	for i, start := range breaks {
		end := len(input)
		if i+1 < len(breaks) {
			end = breaks[i+1]
		}
		lines = append(lines, input[start:end])
	}
	cursorRow = 0
	for i := len(breaks) - 1; i >= 0; i-- {
		if breaks[i] <= cursor {
			cursorRow = i
			break
		}
	}
	for _, cluster := range input[breaks[cursorRow]:cursor] {
		cursorCol += max(1, displaywidth.String(cluster))
	}
	return
}

// maxInputHeight caps how many rows the input box may grow to, leaving room
// for the chat area above it.
func maxInputHeight(h int) int {
	cap := min(6, h/3)
	return max(1, cap)
}

// inputHeight returns how many rows drawInput needs for the current text.
func inputHeight(input []string, cursor, width, screenHeight int) int {
	if width <= 2 {
		return 1
	}
	lines, _, _ := wrapInput(input, cursor, width-2)
	return min(max(1, len(lines)), maxInputHeight(screenHeight))
}
