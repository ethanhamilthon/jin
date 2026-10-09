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
	scroll      *int
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
		if box.scroll != nil {
			*box.scroll = 0
		}
		put(screen, 2, top, box.placeholder, dim)
		if box.focused {
			screen.ShowCursor(2, top)
		}
		return
	}
	renumberImages(box.text)
	lines, cursorRow, cursorCol := wrapInput(box.visible(), box.cursor, width-2)
	start := 0
	if box.scroll != nil {
		start = fitScroll(*box.scroll, cursorRow, len(lines), height)
		*box.scroll = start
	}
	for i := 0; i < height && start+i < len(lines); i++ {
		x := 2
		for _, cluster := range lines[start+i] {
			cells := clusterWidth(cluster)
			if x+cells > width {
				break
			}
			style := textStyle
			if isTokenMark(cluster) {
				style = tokenStyle(box.focused)
			}
			put(screen, x, top+i, shownCluster(cluster), style)
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

// fitScroll keeps the view where it was and moves it only as far as needed
// for the cursor row to stay visible.
func fitScroll(start, cursorRow, total, height int) int {
	start = min(start, max(0, total-height))
	if cursorRow < start {
		return cursorRow
	}
	if cursorRow >= start+height {
		return cursorRow - height + 1
	}
	return start
}
