package ui

import (
	"errors"
	"jin/internal/store"
	"strings"
	"testing"
	"time"
)

func openVoiceCheck(a *app) *voiceCheck {
	sel := &selector{field: true}
	a.sel = sel
	return &voiceCheck{sel: sel}
}

func TestVoiceCheckFailureKeepsFieldOpenWithError(t *testing.T) {
	a, _ := layoutApp(t)
	job := openVoiceCheck(a)
	job.sel.loading = true
	a.receiveVoice(voiceResult{check: job, err: errors.New("401 Unauthorized")})
	if a.sel != job.sel || job.sel.err != "401 Unauthorized" || job.sel.loading || a.cfg.Voice.Ready() {
		t.Fatalf("sel open = %v, err = %q, loading = %v", a.sel == job.sel, job.sel.err, job.sel.loading)
	}
}

func TestVoiceWaveDrawsRecentLevels(t *testing.T) {
	a, screen := layoutApp(t)
	v := &voiceState{session: a.active, recording: true}
	v.levels = []float64{0, 0.5, 1}
	a.voice = v
	a.drawWave(10, 5, 30)
	if got := rowText(screen, 5, 60); !strings.Contains(got, "▅█") {
		t.Errorf("row = %q", got)
	}
}

func TestVoiceCheckSuccessSavesBeforeContinuing(t *testing.T) {
	a, _ := layoutApp(t)
	db, _ := openFoldDB(t)
	a.store = db
	job := openVoiceCheck(a)
	job.cfg = store.Voice{BaseURL: "https://fal.run", APIKey: "test", Model: "fal-ai/wizper"}
	called := false
	job.done = func() { called = true }
	a.receiveVoice(voiceResult{check: job})
	cfg, err := db.LoadConfig()
	if err != nil || cfg.Voice != job.cfg || a.cfg.Voice != job.cfg || a.sel != nil || !called {
		t.Fatalf("saved=%v closed=%v continued=%v err=%v", cfg.Voice == job.cfg, a.sel == nil, called, err)
	}
}

func TestVoiceCheckCancelledDoesNotSave(t *testing.T) {
	a, _ := layoutApp(t)
	db, _ := openFoldDB(t)
	a.store = db
	job := openVoiceCheck(a)
	job.cfg = store.Voice{BaseURL: "https://fal.run", APIKey: "test", Model: "fal-ai/wizper"}
	a.sel = nil
	a.receiveVoice(voiceResult{check: job})
	cfg, err := db.LoadConfig()
	if err != nil || cfg.Voice.Ready() {
		t.Fatalf("cancelled config saved, err=%v", err)
	}
}

func TestVoiceTitleReservesWaveOnNarrowScreen(t *testing.T) {
	a, screen := layoutApp(t)
	a.voice = &voiceState{session: a.active, recording: true, since: time.Now(), levels: []float64{0.5, 1}}
	a.drawVoiceTitle(5, 60)
	if got := rowText(screen, 5, 60); !strings.Contains(got, "▅█") {
		t.Fatalf("wave hidden by title: %q", got)
	}
}
