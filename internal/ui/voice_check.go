package ui

import (
	"context"
	"jin/internal/store"
	"jin/internal/voice"
)

type voiceCheck struct {
	sel  *selector
	cfg  store.Voice
	done func()
}

// checkVoice tries the endpoint with a second of silence and saves the
// settings only when it answers. The last field stays open meanwhile.
func (a *app) checkVoice(v store.Voice, done func()) {
	sel := a.sel
	if sel == nil || sel.loading {
		return
	}
	sel.loading, sel.keepOpen = true, true
	job := &voiceCheck{sel: sel, cfg: v, done: done}
	client := voice.Client{BaseURL: v.BaseURL, APIKey: v.APIKey, Model: v.Model, Language: v.Language}
	go func() {
		ctx, cancel := context.WithTimeout(a.ctx, voiceTimeout)
		defer cancel()
		err := client.Check(ctx)
		select {
		case a.voiceDone <- voiceResult{check: job, err: err}:
		case <-a.ctx.Done():
		}
	}()
}

func (a *app) receiveVoiceCheck(job *voiceCheck, err error) {
	sel := job.sel
	if a.sel != sel {
		return
	}
	sel.loading = false
	if err == nil {
		err = a.store.SaveVoice(job.cfg)
	}
	if err != nil {
		sel.err = err.Error()
		return
	}
	a.cfg.Voice = job.cfg
	a.sel = nil
	if job.done != nil {
		job.done()
	}
}
