package ui

import (
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/color"
)

var update = flag.Bool("update", false, "rewrite the golden screens in testdata")

// paletteMarks names the theme colors in a golden screen, so a test catches
// a cell that changed color, not only one that changed text.
func paletteMarks() map[color.Color]byte {
	return map[color.Color]byte{
		colorBG: '.', colorFG: 'F', colorText: 't', colorMuted: 'm', colorArgument: 'a',
		colorDetail: 'd', colorDim: 'i', colorBorder: 'b', colorRaised: 'r', colorBlue: 'S',
		colorOnBlue: 'o', colorWhite: 'W', colorBlueFG: 'B', colorGreen: 'G', colorAmber: 'A',
		colorRed: 'R', colorPurple: 'P', colorPink: 'K', colorTeal: 'T',
	}
}

// snapshot is the screen as text, then the foreground and background of
// every cell as palette letters ('?' for a color outside the palette, such
// as a gradient step).
func snapshot(screen tcell.Screen) string {
	w, h := screen.Size()
	marks := paletteMarks()
	mark := func(c color.Color) byte {
		if m, ok := marks[c]; ok {
			return m
		}
		return '?'
	}
	var text, fg, bg strings.Builder
	for y := range h {
		for x := 0; x < w; {
			str, style, width := screen.Get(x, y)
			if str == "" {
				str = " "
			}
			text.WriteString(str)
			for range max(1, width) {
				fg.WriteByte(mark(style.GetForeground()))
				bg.WriteByte(mark(style.GetBackground()))
			}
			x += max(1, width)
		}
		text.WriteByte('\n')
		fg.WriteByte('\n')
		bg.WriteByte('\n')
	}
	return text.String() + "\n-- foreground --\n" + fg.String() + "\n-- background --\n" + bg.String() + "\n-- palette --\n" + paletteText(marks)
}

// paletteText lists each palette letter with its color, so a changed theme
// color shows in the golden file too.
func paletteText(marks map[color.Color]byte) string {
	var lines []string
	for c, m := range marks {
		r, g, b := c.RGB()
		lines = append(lines, string(m)+" #"+strconv.FormatInt(int64(r)<<16|int64(g)<<8|int64(b)+1<<24, 16)[1:])
	}
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

// golden compares the screen with testdata/<name>.golden; go test -update
// writes it instead.
func golden(t *testing.T, screen tcell.Screen, name string) {
	t.Helper()
	got := snapshot(screen)
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test ./internal/ui -run Golden -update)", err)
	}
	if got != string(want) {
		t.Errorf("screen %s changed; check it and run go test ./internal/ui -run Golden -update\n%s", name, firstDiff(string(want), got))
	}
}

func firstDiff(want, got string) string {
	wl, gl := strings.Split(want, "\n"), strings.Split(got, "\n")
	for i := 0; i < max(len(wl), len(gl)); i++ {
		var w, g string
		if i < len(wl) {
			w = wl[i]
		}
		if i < len(gl) {
			g = gl[i]
		}
		if w != g {
			return "line " + strconv.Itoa(i+1) + "\nwant: " + w + "\ngot:  " + g
		}
	}
	return ""
}
