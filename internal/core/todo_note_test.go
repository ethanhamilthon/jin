package core

import (
	"testing"

	"jin/internal/todo"
)

func TestTodoEditedBlockIsStripped(t *testing.T) {
	block := TodoEditedBlock([]todo.Item{{Text: "a", Status: todo.Done}})
	if got := StripTodoEdited(block + "hello"); got != "hello" {
		t.Fatalf("got %q", got)
	}
	if got := StripTodoEdited("plain"); got != "plain" {
		t.Fatalf("got %q", got)
	}
}
