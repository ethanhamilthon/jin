//go:build unix

package async

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestTaskStdinIsClosedByDefault(t *testing.T) {
	db := startTestDaemon(t)
	cwd, _ := os.Getwd()
	start := time.Now()
	id, err := Start("test", "s", "cat; echo eof", false)
	if err != nil {
		t.Fatal(err)
	}
	event := waitEvent(t, db, cwd)
	if !strings.Contains(event.Text, `status="done"`) || !strings.Contains(event.Text, "eof") {
		t.Errorf("event = %s", event.Text)
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("cat waited for stdin: %v", time.Since(start))
	}
	if err := Input(id, "x", false); err == nil {
		t.Error("input to a finished task must fail")
	}
}

func TestInputWithoutStdinFails(t *testing.T) {
	startTestDaemon(t)
	id, err := Start("test", "s", "sleep 30", false)
	if err != nil {
		t.Fatal(err)
	}
	defer Stop(id, StoppedByAgent)
	if err := Input(id, "x", false); err == nil || !strings.Contains(err.Error(), "--stdin") {
		t.Errorf("input without --stdin: %v", err)
	}
}

func TestInputToAChildThatNeverReadsGivesUp(t *testing.T) {
	startTestDaemon(t)
	inputTimeout = 300 * time.Millisecond
	t.Cleanup(func() { inputTimeout = 5 * time.Second })
	id, err := Start("test", "s", "sleep 30", true)
	if err != nil {
		t.Fatal(err)
	}
	defer Stop(id, StoppedByAgent)
	big := strings.Repeat("x", 1<<20)
	start := time.Now()
	if err := Input(id, big, false); err == nil || !strings.Contains(err.Error(), "does not read") {
		t.Errorf("input = %v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Errorf("input hung: %v", time.Since(start))
	}
	if err := Input(id, "small", false); err == nil {
		t.Error("a full pipe must refuse more input")
	}
}
