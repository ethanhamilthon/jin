package core

import (
	"strings"

	"jin/internal/wire"
)

// todoEditedOpen and todoEditedClose wrap a note that older versions of jin
// put in front of a prompt; old sessions still carry it.
const (
	todoEditedOpen  = wire.TodoEditedOpen
	todoEditedClose = wire.TodoEditedClose
)

// StripTodoEdited removes that note, leaving what the user typed. A
// prompts block that follows it is left for prompts.Strip.
func StripTodoEdited(content string) string {
	if !strings.HasPrefix(content, todoEditedOpen) {
		return content
	}
	if _, rest, ok := strings.Cut(content, todoEditedClose+"\n\n"); ok {
		return rest
	}
	return content
}

const (
	undoOpen  = wire.UndoOpen
	undoClose = wire.UndoClose
)

// UndoBlock tells the model the user reverted its file changes. It goes in
// front of the next user prompt.
func UndoBlock(paths []string) string { return wire.UndoBlock(paths) }

// StripUndo undoes UndoBlock, leaving what follows it.
func StripUndo(content string) string {
	for strings.HasPrefix(content, undoOpen) {
		_, rest, ok := strings.Cut(content, undoClose+"\n\n")
		if !ok {
			return content
		}
		content = rest
	}
	return content
}

// StripNotes removes every note jin puts in front of a user prompt, in any
// order, leaving the prompt itself.
func StripNotes(content string) string {
	for {
		next := StripRefreshed(StripTodoEdited(StripUndo(content)))
		if next == content {
			return content
		}
		content = next
	}
}
