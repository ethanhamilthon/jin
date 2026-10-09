package core

import "testing"

func TestOldTodoEditedNoteIsStripped(t *testing.T) {
	note := "<todo-edited>The user edited the todo list.</todo-edited>\n\n"
	if got := StripTodoEdited(note + "hello"); got != "hello" {
		t.Fatalf("got %q", got)
	}
	if got := StripTodoEdited("plain"); got != "plain" {
		t.Fatalf("got %q", got)
	}
}
