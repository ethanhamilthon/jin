package ui

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

func selectedText(rows []chatRow, sel textSelection) string {
	if !sel.active || sel.startRow < 0 || sel.endRow < 0 || sel.startRow >= len(rows) || sel.endRow >= len(rows) {
		return ""
	}
	firstRow, firstCol := sel.startRow, sel.startCol
	lastRow, lastCol := sel.endRow, sel.endCol
	if firstRow > lastRow || (firstRow == lastRow && firstCol > lastCol) {
		firstRow, lastRow = lastRow, firstRow
		firstCol, lastCol = lastCol, firstCol
	}
	if firstRow == lastRow && firstCol == lastCol {
		return ""
	}
	var result strings.Builder
	for row := firstRow; row <= lastRow; row++ {
		if row > firstRow {
			result.WriteByte('\n')
		}
		text := rows[row].text
		origin := 2
		from, to := origin, int(^uint(0)>>1)
		if row == firstRow {
			from = firstCol
		}
		if row == lastRow {
			to = lastCol
		}
		col := origin
		graphemes := displaywidth.StringGraphemes(text)
		for graphemes.Next() {
			cluster, width := graphemes.Value(), graphemes.Width()
			if width > 0 && col+width > from && col < to {
				result.WriteString(cluster)
			}
			col += width
		}
	}
	return result.String()
}

// copySelection writes text to the system clipboard with pbcopy, wl-copy or
// xclip when present, and falls back to the terminal (OSC 52).
func copySelection(screen tcell.Screen, text string) {
	if args := copyCommand(); args != nil {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = strings.NewReader(text)
		if cmd.Run() == nil {
			return
		}
	}
	screen.SetClipboard([]byte(text))
}

func copyCommand() []string {
	switch {
	case runtime.GOOS == "darwin":
		return []string{"pbcopy"}
	case os.Getenv("WAYLAND_DISPLAY") != "" && onPath("wl-copy"):
		return []string{"wl-copy"}
	case os.Getenv("DISPLAY") != "" && onPath("xclip"):
		return []string{"xclip", "-selection", "clipboard"}
	}
	return nil
}
