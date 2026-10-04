package ui

import (
	"errors"
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
)

func TestVoiceEnterWaitsThenSendsOnce(t *testing.T) {
	a, v := voiceApp(t)
	a.active.agent = core.NewAgent(nil, "", a.registry)
	db, _ := openFoldDB(t)
	a.store, a.active.store, a.active.path = db, db, "p"
	v.finals, v.recording = 1, false
	a.voiceKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	a.voiceKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	if len(a.active.pending) != 0 || a.voice == nil {
		t.Fatal("Enter must wait for transcription")
	}
	a.receiveVoice(voiceResult{owner: v, text: "the bug"})
	a.receiveVoice(voiceResult{owner: v, text: "duplicate"})
	if a.voice != nil || draftText(a) != "" || len(a.active.pending) != 1 {
		t.Fatalf("voice=%v draft=%q pending=%d", a.voice != nil, draftText(a), len(a.active.pending))
	}
	if got := a.active.pending[0].Prompt; got != "fix the bug" {
		t.Fatalf("prompt = %q", got)
	}
}

func TestVoiceEnterErrorKeepsDraft(t *testing.T) {
	a, v := voiceApp(t)
	v.finals, v.recording = 1, false
	a.voiceKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	a.receiveVoice(voiceResult{owner: v, err: errors.New("boom")})
	if a.voice != nil || draftText(a) != "fix" || len(a.active.pending) != 0 {
		t.Fatalf("voice=%v draft=%q pending=%d", a.voice != nil, draftText(a), len(a.active.pending))
	}
}

func TestVoiceEmptyTranscriptDoesNotSendExistingDraft(t *testing.T) {
	for _, pending := range []int{0, 1} {
		a, v := voiceApp(t)
		v.finals, v.recording = pending, false
		a.voiceKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
		if pending != 0 {
			a.receiveVoice(voiceResult{owner: v, text: "  "})
		}
		if a.voice != nil || draftText(a) != "fix" || len(a.active.pending) != 0 {
			t.Fatalf("voice=%v draft=%q pending=%d", a.voice != nil, draftText(a), len(a.active.pending))
		}
	}
}

func TestVoiceEnterRefusalPreservesTranscript(t *testing.T) {
	a, v := voiceApp(t)
	a.active.readOnlyPID = 123
	v.finals, v.recording = 1, false
	a.voiceKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	a.receiveVoice(voiceResult{owner: v, text: "the bug"})
	if a.voice != nil || draftText(a) != "fix the bug" || len(a.active.pending) != 0 {
		t.Fatalf("voice=%v draft=%q pending=%d", a.voice != nil, draftText(a), len(a.active.pending))
	}
}

func TestVoiceEnterCancelledIgnoresLateResult(t *testing.T) {
	a, v := voiceApp(t)
	v.finals, v.recording = 1, false
	a.voiceKey(tcell.NewEventKey(tcell.KeyEnter, "", tcell.ModNone))
	a.cancelVoice()
	a.receiveVoice(voiceResult{owner: v, text: "the bug"})
	if draftText(a) != "fix" || len(a.active.pending) != 0 {
		t.Fatalf("draft=%q pending=%d", draftText(a), len(a.active.pending))
	}
}
