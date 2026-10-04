package ui

import (
	"fmt"
	"time"
)

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
	case v.err != "":
		title = "Recording " + clock + " · " + v.err
	}
	put(a.screen, 2, y, " "+truncate(title, w-6)+" ", accent.Bold(true))
}
