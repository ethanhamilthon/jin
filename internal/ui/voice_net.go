package ui

import (
	"context"
	"strings"
	"time"

	"jin/internal/voice"
)

const (
	voiceEvery   = 1500 * time.Millisecond
	voiceTimeout = 30 * time.Second
)

type voiceResult struct {
	owner *voiceState
	id    int
	final bool
	text  string
	err   error
}

// transcribe sends audio in the background. A final request is for audio
// that is complete; a live one is a look at a segment that is still running.
func (a *app) transcribe(pcm []byte, final bool) {
	v := a.voice
	v.seq++
	id := v.seq
	cfg := a.cfg.Voice
	client := voice.Client{BaseURL: cfg.BaseURL, APIKey: cfg.APIKey, Model: cfg.Model, Language: cfg.Language}
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, voiceTimeout)
		defer cancel()
		text, err := client.Transcribe(ctx, pcm)
		select {
		case a.voiceDone <- voiceResult{owner: v, id: id, final: final, text: text, err: err}:
		case <-a.ctx.Done():
		}
	}()
}

// voiceTick sends what was said so far, so the draft follows the speech.
func (a *app) voiceTick() {
	v := a.voice
	if v == nil || !v.recording || v.busy || time.Since(v.lastSent) < voiceEvery {
		return
	}
	pcm := v.rec.Snapshot()
	if len(pcm) == v.sentLen || voice.Speechless(pcm) {
		return
	}
	v.busy, v.lastSent, v.sentLen = true, time.Now(), len(pcm)
	a.transcribe(pcm, false)
}

func (a *app) receiveVoice(r voiceResult) {
	v := a.voice
	if v == nil || r.owner != v {
		return
	}
	if r.final {
		v.finals--
	} else {
		v.busy = false
		if r.id != v.seq {
			return
		}
	}
	switch {
	case r.err != nil && r.final:
		v.live = ""
		a.report(r.err)
	case r.err != nil:
		v.err = r.err.Error()
		return
	case r.final:
		v.base, v.live, v.err = strings.TrimSpace(v.base+" "+r.text), "", ""
	case !v.recording:
		return
	default:
		v.live, v.err = r.text, ""
	}
	a.showVoice()
	a.leaveVoice()
}
