package ui

import (
	"time"

	"github.com/gdamore/tcell/v3"

	"jin/internal/voice"
)

// voiceKey handles a key while /voice listens: Space pauses and resumes,
// Enter keeps the text in the draft, Esc drops it.
func (a *app) voiceKey(ev *tcell.EventKey) {
	v := a.voice
	switch {
	case a.pasting:
	case ev.Key() == tcell.KeyEscape:
		a.cancelVoice()
	case ev.Key() == tcell.KeyEnter:
		if v.exit {
			return
		}
		v.exit = true
		a.pauseVoice()
		a.leaveVoice()
	case ev.Key() == tcell.KeyRune && ev.Str() == " ":
		if v.recording {
			a.pauseVoice()
		} else if v.finals == 0 && !v.exit {
			a.resumeVoice()
		}
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

func (a *app) resumeVoice() {
	v := a.voice
	if err := v.rec.Start(); err != nil {
		v.exit = true
		a.leaveVoice()
		a.report(err)
		return
	}
	v.recording, v.since = true, time.Now()
}
