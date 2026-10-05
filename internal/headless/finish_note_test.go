package headless

import (
	"testing"
)

func TestReplyAfterTasksNoteJoinsTheAnswer(t *testing.T) {
	o := &outcome{}
	o.answer = joinAnswer(o, "Done.")
	o.afterNote = true
	if got := joinAnswer(o, "The server keeps running."); got != "Done.\n\nThe server keeps running." {
		t.Fatalf("answer = %q", got)
	}
	if o.afterNote {
		t.Fatal("the note joins only one reply")
	}
}
