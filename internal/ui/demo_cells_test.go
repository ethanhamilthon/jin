package ui

import "github.com/gdamore/tcell/v3/color"

type demoCell struct {
	Text string   `json:"text"`
	FG   [3]int32 `json:"fg"`
	BG   [3]int32 `json:"bg"`
}

type demoFrame struct {
	Label  string     `json:"label"`
	Delay  int        `json:"delay"`
	Width  int        `json:"width"`
	Height int        `json:"height"`
	Cells  []demoCell `json:"cells"`
}

func demoRGB(c color.Color, fallback color.Color) [3]int32 {
	r, g, b := c.RGB()
	if r < 0 || g < 0 || b < 0 {
		r, g, b = fallback.RGB()
	}
	return [3]int32{r, g, b}
}

func captureDemoFrame(a *app, label string, delay int) demoFrame {
	w, h := a.screen.Size()
	frame := demoFrame{Label: label, Delay: delay, Width: w, Height: h}
	for y := range h {
		for x := range w {
			text, style, _ := a.screen.Get(x, y)
			frame.Cells = append(frame.Cells, demoCell{Text: text,
				FG: demoRGB(style.GetForeground(), colorFG), BG: demoRGB(style.GetBackground(), colorBG)})
		}
	}
	return frame
}
