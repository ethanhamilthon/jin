package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

var spinnerFrames = []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

type inputBox struct {
	text        []string
	cursor      int
	prefix      string
	prefixStyle tcell.Style
	placeholder string
	focused     bool
	secret      bool
}

func (box inputBox) visible() []string {
	if !box.secret {
		return box.text
	}
	masked := make([]string, len(box.text))
	for i := range masked {
		masked[i] = "•"
	}
	return masked
}

func drawInput(screen tcell.Screen, box inputBox, top, height, width int) {
	screen.HideCursor()
	if height <= 0 || width <= 2 {
		return
	}
	put(screen, 0, top, box.prefix, box.prefixStyle)
	textStyle := base
	if !box.focused {
		textStyle = muted
	}
	if len(box.text) == 0 {
		put(screen, 2, top, box.placeholder, dim)
		if box.focused {
			screen.ShowCursor(2, top)
		}
		return
	}
	lines, cursorRow, cursorCol := wrapInput(box.visible(), box.cursor, width-2)
	start := 0
	if maxStart := len(lines) - height; maxStart > 0 {
		start = min(maxStart, max(0, cursorRow-height+1))
	}
	for i := 0; i < height && start+i < len(lines); i++ {
		x := 2
		for _, cluster := range lines[start+i] {
			cells := max(1, displaywidth.String(cluster))
			if x+cells > width {
				break
			}
			put(screen, x, top+i, cluster, textStyle)
			x += cells
		}
	}
	if box.focused {
		screen.ShowCursor(min(width-1, 2+cursorCol), top+cursorRow-start)
	}
}

func clusters(text string) []string {
	var out []string
	graphemes := displaywidth.StringGraphemes(text)
	for graphemes.Next() {
		out = append(out, graphemes.Value())
	}
	return out
}
