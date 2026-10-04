package ui

import (
	"errors"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/voice"
)

func voiceApp(t *testing.T) (*app, *voiceState) {
	a, _ := layoutApp(t)
	a.voiceDone = make(chan voiceResult, 8)
	s := a.active
	s.input = clusters("fix")
	s.cursor = len(s.input)
	v := &voiceState{rec: &voice.Recorder{}, session: s, start: s.cursor, lead: " ", recording: true}
	a.voice = v
	return a, v
}

func draftText(a *app) string { return strings.Join(a.active.input, "") }

func TestVoiceFinalKeepsTextAndLeavesOnEnter(t *testing.T) {
	a, v := voiceApp(t)
	v.finals, v.recording, v.exit = 1, false, true
	a.receiveVoice(voiceResult{owner: v, text: "the bug"})
	if a.voice != nil {
		t.Fatal("voice mode should end after the last answer")
	}
	if got := draftText(a); got != "fix the bug" || a.active.cursor != len(a.active.input) {
		t.Fatalf("draft = %q, cursor = %d", got, a.active.cursor)
	}
}

func TestVoicePausedFinalsAccumulate(t *testing.T) {
	a, v := voiceApp(t)
	v.recording = false
	v.finals = 1
	a.receiveVoice(voiceResult{owner: v, text: "one"})
	v.finals = 1
	a.receiveVoice(voiceResult{owner: v, text: "two"})
	if got := draftText(a); got != "fix one two" || a.voice == nil {
		t.Fatalf("draft = %q, voice open = %v", got, a.voice != nil)
	}
}

func TestVoiceFinalErrorIsReported(t *testing.T) {
	a, v := voiceApp(t)
	v.finals, v.recording = 1, false
	a.receiveVoice(voiceResult{owner: v, err: errors.New("boom")})
	last := a.active.history[len(a.active.history)-1]
	if !strings.Contains(last.text, "boom") {
		t.Fatalf("last entry = %q", last.text)
	}
}

func TestVoiceEscDropsTranscript(t *testing.T) {
	a, v := voiceApp(t)
	v.finals, v.recording = 1, false
	a.receiveVoice(voiceResult{owner: v, text: "gone"})
	a.cancelVoice()
	if a.voice != nil || draftText(a) != "fix" {
		t.Fatalf("voice open = %v, draft = %q", a.voice != nil, draftText(a))
	}
}

func TestVoicePasteDoesNotDriveKeys(t *testing.T) {
	a, v := voiceApp(t)
	a.handleEvent(tcell.NewEventPaste(true))
	press(a, tcell.KeyEnter)
	press(a, tcell.KeyEscape)
	a.handleEvent(tcell.NewEventPaste(false))
	if a.voice != v || v.exit || draftText(a) != "fix" {
		t.Fatalf("voice open = %v, exit = %v, draft = %q", a.voice == v, v.exit, draftText(a))
	}
}
