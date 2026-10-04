package ui

import (
	"github.com/clipperhouse/displaywidth"
	"github.com/gdamore/tcell/v3"
)

type clippedScreen struct {
	tcell.Screen
	paneRect
}

func (s *clippedScreen) Size() (int, int) { return s.w, s.h }

func (s *clippedScreen) Get(x, y int) (string, tcell.Style, int) {
	if x < 0 || y < 0 || x >= s.w || y >= s.h {
		return "", tcell.StyleDefault, 0
	}
	return s.Screen.Get(s.x+x, s.y+y)
}

func (s *clippedScreen) Put(x, y int, text string, style tcell.Style) (string, int) {
	g := displaywidth.StringGraphemes(text)
	if !g.Next() || x < 0 || y < 0 || y >= s.h || x+g.Width() > s.w {
		return text, 0
	}
	cluster := g.Value()
	rest := text[len(cluster):]
	_, width := s.Screen.Put(s.x+x, s.y+y, cluster, style)
	return rest, width
}

func (s *clippedScreen) PutStr(x, y int, text string) {
	s.PutStrStyled(x, y, text, tcell.StyleDefault)
}

func (s *clippedScreen) PutStrStyled(x, y int, text string, style tcell.Style) {
	if y < 0 || y >= s.h {
		return
	}
	g := displaywidth.StringGraphemes(text)
	for g.Next() {
		cluster, width := g.Value(), g.Width()
		if x >= 0 && x+width <= s.w {
			s.Screen.PutStrStyled(s.x+x, s.y+y, cluster, style)
		}
		x += width
		if x >= s.w {
			return
		}
	}
}

func (s *clippedScreen) SetContent(x, y int, primary rune, combining []rune, style tcell.Style) {
	width := displaywidth.String(string(primary))
	if x < 0 || y < 0 || x+width > s.w || y >= s.h {
		return
	}
	s.Screen.SetContent(s.x+x, s.y+y, primary, combining, style)
}

func (s *clippedScreen) Fill(r rune, style tcell.Style) {
	s.FillArea(0, 0, s.w, s.h, r, style)
}

func (s *clippedScreen) FillArea(x, y, width, height int, r rune, style tcell.Style) {
	for row := max(0, y); row < min(s.h, y+height); row++ {
		for col := max(0, x); col < min(s.w, x+width); col++ {
			s.SetContent(col, row, r, nil, style)
		}
	}
}

func (s *clippedScreen) ShowCursor(x, y int) {
	if x < 0 || y < 0 || x >= s.w || y >= s.h {
		s.Screen.HideCursor()
		return
	}
	s.Screen.ShowCursor(s.x+x, s.y+y)
}
