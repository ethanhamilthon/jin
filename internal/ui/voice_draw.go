package ui

import (
	"fmt"
	"time"

	"github.com/clipperhouse/displaywidth"
)

var waveBars = []string{"▁", "▂", "▃", "▄", "▅", "▆", "▇", "█"}

// voiceBox is the input row while /voice listens: the draft with the
// transcript in it, marked by a dot that is red while the microphone is on.
func (a *app) voiceBox() inputBox {
	v, s := a.voice, a.voice.session
	prefix, style := "● ", errorStyle.Bold(true)
	if !v.recording {
		prefix, style = "‖ ", dim
	}
	return inputBox{text: s.input, cursor: s.cursor, prefix: prefix, prefixStyle: style,
		placeholder: "Speak...", focused: true, scroll: &s.inputTop}
}

// drawVoiceTitle writes the status on the rule above the input and runs the
// sound wave along the rest of it, newest sound at the right end.
func (a *app) drawVoiceTitle(y, w int) {
	v := a.voice
	heard := v.heard
	if v.recording {
		heard += time.Since(v.since)
	}
	clock := fmt.Sprintf("%d:%02d", int(heard.Minutes()), int(heard.Seconds())%60)
	title := "Recording " + clock + " · Space pause · Enter keep · Esc drop"
	switch {
	case v.finals > 0:
		title = "Transcribing... · Enter keep · Esc drop"
	case !v.recording:
		title = "Paused " + clock + " · Space resume · Enter keep · Esc drop"
	}
	limit := w - 6
	if v.recording {
		limit -= min(16, w/4)
		if displaywidth.String(title) > limit {
			title = "Rec " + clock + " · Space · Enter · Esc"
		}
	}
	title = truncate(title, max(1, limit))
	put(a.screen, 2, y, " "+title+" ", accent.Bold(true))
	if v.recording {
		a.drawWave(2+displaywidth.String(title)+3, y, w-1)
	}
}

func (a *app) drawWave(from, y, to int) {
	v := a.voice
	n := to - from
	if n <= 0 {
		return
	}
	levels := v.levels
	if len(levels) > n {
		levels = levels[len(levels)-n:]
	}
	x := to - len(levels)
	style := base.Foreground(colorRed)
	for i, level := range levels {
		bar := int(level * float64(len(waveBars)))
		if bar >= len(waveBars) {
			bar = len(waveBars) - 1
		}
		if level < 0.02 {
			put(a.screen, x+i, y, "─", border)
			continue
		}
		put(a.screen, x+i, y, waveBars[bar], style)
	}
}
