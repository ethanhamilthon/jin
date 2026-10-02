package ui

import (
	"errors"
	"testing"
	"time"
)

// While an editor or a /tui program owns the terminal, the event loop must
// keep taking what the background sources deliver.
func TestServeUntilKeepsHandlingEventsWhileAProgramRuns(t *testing.T) {
	a := &app{
		updates:      make(chan taggedUpdate, 1),
		loads:        make(chan loadResult, 1),
		bashDone:     make(chan bashResult, 1),
		modelsLoaded: make(chan modelsResult, 1),
		asyncs:       make(chan asyncBatch, 1),
		sessions:     map[string]*chatSession{},
		asyncRunning: map[string]int{},
	}
	done := make(chan error, 1)
	finished := make(chan error, 1)
	go func() { finished <- a.serveUntil(done) }()

	a.asyncs <- asyncBatch{running: map[string]int{"s1": 2}}
	deadline := time.After(2 * time.Second)
	for {
		// The loop took the batch when the channel is empty again.
		if len(a.asyncs) == 0 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("the batch was not taken while the program ran")
		case <-time.After(10 * time.Millisecond):
		}
	}
	want := errors.New("editor exited with 1")
	done <- want
	select {
	case got := <-finished:
		if got != want {
			t.Errorf("serveUntil returned %v, want the program's error", got)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("serveUntil did not return after the program ended")
	}
}
