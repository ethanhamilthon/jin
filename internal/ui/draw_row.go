package ui

import (
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

func drawRow(screen tcell.Screen, y, w int, row chatRow, frame, glow int) {
	if row.hasFill {
		left, right := 1, 3+displaywidth.String(row.text)
		if row.fillWide {
			left, right = 0, w
		}
		for x := max(0, left); x < min(w, right); x++ {
			put(screen, x, y, " ", row.fill)
		}
	}
	if len(row.spans) == 0 {
		style := rowStyle(row.kind)
		if row.hasFill {
			style = row.fill
		}
		put(screen, 2, y, row.text, style)
		return
	}
	x := 2
	for _, span := range row.spans {
		text := span.text
		if span.spin {
			text = spinnerFrames[frame%len(spinnerFrames)]
		}
		if span.shimmer {
			x = drawShimmer(screen, x, y, text, span.style, glow)
			continue
		}
		put(screen, x, y, text, span.style)
		x += displaywidth.String(text)
	}
}

// rule draws a horizontal line, optionally carrying a title: "── Title ────".
func rule(screen tcell.Screen, y, w int, title string) {
	line := strings.Repeat("─", w)
	put(screen, 0, y, line, border)
	if title != "" {
		put(screen, 2, y, " "+truncate(title, w-6)+" ", accent.Bold(true))
	}
}
