package ui

import (
	"strings"
	"time"

	"github.com/gdamore/tcell/v3"

	"jin/internal/voice"
)

// voiceKey handles a key while /voice listens: Space keeps the draft,
// Enter sends the transcript, Esc drops it.
func (a *app) voiceKey(ev *tcell.EventKey) {
	v := a.voice
	switch {
	case a.pasting || !ev.Pressed():
	case ev.Key() == tcell.KeyEscape:
		a.cancelVoice()
	case ev.Key() == tcell.KeyEnter || ev.Key() == tcell.KeyRune && ev.Str() == " ":
		if v.exit {
			return
		}
		v.exit, v.send = true, ev.Key() == tcell.KeyEnter
		a.pauseVoice()
		a.leaveVoice()
	}
}

func (a *app) pauseVoice() {
	v := a.voice
	if !v.recording {
		return
	}
	v.recording = false
	v.heard += time.Since(v.since)
	_ = v.rec.Stop()
	pcm := v.rec.Take()
	if !voice.Speechless(pcm) {
		a.transcribe(pcm)
	}
}

// leaveVoice waits for transcription before keeping or sending the draft.
func (a *app) leaveVoice() {
	if v := a.voice; v != nil && v.exit && v.finals == 0 {
		a.showVoice()
		a.closeVoice()
		if v.send && strings.TrimSpace(v.base) != "" && a.active == v.session {
			a.typeKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
		}
	}
}
