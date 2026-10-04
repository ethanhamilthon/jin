package ui

import (
	"strings"
	"time"
	"unicode"

	"jin/internal/voice"
)

// voiceState is the /voice mode. Its text lives in the draft of one session,
// between start and start+n, and is rewritten as the transcript grows.
type voiceState struct {
	rec     *voice.Recorder
	session *chatSession
	start   int
	n       int
	lead    string

	// base is the text of finished segments, live the text of the running one.
	base, live string
	recording  bool
	exit       bool
	busy       bool
	finals     int
	seq        int
	err        string

	since    time.Time
	heard    time.Duration
	lastSent time.Time
	sentLen  int
}

// startVoice begins listening and shows the words in the draft as they come.
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
	text := strings.Join(strings.Fields(v.base+" "+v.live), " ")
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
	a.voice.base, a.voice.live = "", ""
	a.showVoice()
	a.closeVoice()
}

// leaveVoice keeps the transcript in the draft once no answer is awaited.
func (a *app) leaveVoice() {
	if v := a.voice; v != nil && v.exit && v.finals == 0 {
		a.showVoice()
		a.closeVoice()
	}
}
