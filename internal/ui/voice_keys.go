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
	if voice.Speechless(pcm) {
		v.live = ""
		a.showVoice()
		return
	}
	v.finals++
	a.transcribe(pcm, true)
}

func (a *app) resumeVoice() {
	v := a.voice
	if err := v.rec.Start(); err != nil {
		a.cancelVoice()
		a.report(err)
		return
	}
	v.recording, v.since, v.lastSent, v.sentLen, v.err = true, time.Now(), time.Now(), 0, ""
}
