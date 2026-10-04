package ui

import (
	"strings"
	"time"
	"unicode"

	"jin/internal/voice"
)

// voiceState is the /voice mode. Its text lives in the draft of one session,
// between start and start+n, and grows with each transcribed segment.
type voiceState struct {
	rec     *voice.Recorder
	session *chatSession
	start   int
	n       int
	lead    string

	// base is the text of the transcribed segments.
	base      string
	recording bool
	exit      bool
	send      bool
	finals    int

	levels []float64
	since  time.Time
	heard  time.Duration
}

// startVoice begins listening; the words reach the draft when a segment ends.
func (a *app) startVoice() {
	if !a.cfg.Voice.Ready() {
		a.openVoiceProvider(a.startVoice)
		return
	}
	rec, err := voice.NewRecorder()
	if err == nil {
		err = rec.Start()
		if err != nil {
			rec.Close()
		}
	}
	if err != nil {
		a.report(err)
		return
	}
	s := a.active
	v := &voiceState{rec: rec, session: s, start: s.cursor, recording: true, since: time.Now()}
	if s.cursor > 0 && !isBlank(s.input[s.cursor-1]) {
		v.lead = " "
	}
	a.voice = v
}

func isBlank(cluster string) bool {
	return strings.IndexFunc(cluster, func(r rune) bool { return !unicode.IsSpace(r) }) < 0
}

// showVoice rewrites the span of the draft that holds the transcript.
func (a *app) showVoice() {
	v := a.voice
	s := v.session
	text := strings.Join(strings.Fields(v.base), " ")
	var words []string
	if text != "" {
		words = clusters(v.lead + text)
	}
	end := min(v.start+v.n, len(s.input))
	replaceRange(&s.input, &s.cursor, min(v.start, end), end, words...)
	v.n = len(words)
}

func (a *app) closeVoice() {
	if v := a.voice; v != nil {
		v.rec.Close()
		a.voice = nil
	}
}

// cancelVoice drops the transcript and leaves the mode.
func (a *app) cancelVoice() {
	if a.voice == nil {
		return
	}
	a.voice.base = ""
	a.showVoice()
	a.closeVoice()
}
