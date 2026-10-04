package ui

import (
	"errors"
	"strings"
	"testing"

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

func TestVoiceLiveTextReplacesItself(t *testing.T) {
	a, v := voiceApp(t)
	v.seq = 1
	a.receiveVoice(voiceResult{owner: v, id: 1, text: "this"})
	v.seq = 2
	a.receiveVoice(voiceResult{owner: v, id: 2, text: "this bug"})
	if got := draftText(a); got != "fix this bug" {
		t.Fatalf("draft = %q", got)
	}
}

func TestVoiceStaleLiveResultIsDropped(t *testing.T) {
	a, v := voiceApp(t)
	v.seq = 5
	a.receiveVoice(voiceResult{owner: v, id: 4, text: "old"})
	if got := draftText(a); got != "fix" {
		t.Fatalf("draft = %q", got)
	}
}

func TestVoiceFinalKeepsTextAndLeavesOnEnter(t *testing.T) {
	a, v := voiceApp(t)
	v.finals, v.recording, v.exit = 1, false, true
	a.receiveVoice(voiceResult{owner: v, final: true, text: "the bug"})
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
	a.receiveVoice(voiceResult{owner: v, final: true, text: "one"})
	v.finals = 1
	a.receiveVoice(voiceResult{owner: v, final: true, text: "two"})
	if got := draftText(a); got != "fix one two" || a.voice == nil {
		t.Fatalf("draft = %q, voice open = %v", got, a.voice != nil)
	}
}

func TestVoiceFinalErrorIsReported(t *testing.T) {
	a, v := voiceApp(t)
	v.finals, v.recording = 1, false
	a.receiveVoice(voiceResult{owner: v, final: true, err: errors.New("boom")})
	last := a.active.history[len(a.active.history)-1]
	if !strings.Contains(last.text, "boom") {
		t.Fatalf("last entry = %q", last.text)
	}
}

func TestVoiceEscDropsTranscript(t *testing.T) {
	a, v := voiceApp(t)
	v.seq = 1
	a.receiveVoice(voiceResult{owner: v, id: 1, text: "gone"})
	a.cancelVoice()
	if a.voice != nil || draftText(a) != "fix" {
		t.Fatalf("voice open = %v, draft = %q", a.voice != nil, draftText(a))
	}
}
