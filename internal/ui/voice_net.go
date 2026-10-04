package ui

import (
	"context"
	"strings"
	"time"

	"jin/internal/voice"
)

const (
	maxWave      = 200
	voiceTimeout = 60 * time.Second
)

type voiceResult struct {
	check *voiceCheck
	owner *voiceState
	text  string
	err   error
}

// transcribe sends one finished segment in the background.
func (a *app) transcribe(pcm []byte) {
	v := a.voice
	v.finals++
	cfg := a.cfg.Voice
	client := voice.Client{BaseURL: cfg.BaseURL, APIKey: cfg.APIKey, Model: cfg.Model, Language: cfg.Language}
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, voiceTimeout)
		defer cancel()
		text, err := client.Transcribe(ctx, pcm)
		select {
		case a.voiceDone <- voiceResult{owner: v, text: text, err: err}:
		case <-a.ctx.Done():
		}
	}()
}

// voiceTick follows the microphone loudness for the wave. Nothing is sent
// until the segment ends.
func (a *app) voiceTick() {
	v := a.voice
	if v == nil || !v.recording {
		return
	}
	v.levels = append(v.levels, v.rec.Level())
	if len(v.levels) > maxWave {
		v.levels = v.levels[len(v.levels)-maxWave:]
	}
}

func (a *app) receiveVoice(r voiceResult) {
	if r.check != nil {
		a.receiveVoiceCheck(r.check, r.err)
		return
	}
	v := a.voice
	if v == nil || r.owner != v {
		return
	}
	v.finals--
	if r.err != nil {
		v.send = false
		a.report(r.err)
	} else {
		v.base = strings.TrimSpace(v.base + " " + r.text)
	}
	a.showVoice()
	a.leaveVoice()
}
