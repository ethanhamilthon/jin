package core

import (
	"strings"

	"jin/internal/todo"
)

const (
	todoEditedOpen  = "<todo-edited>"
	todoEditedClose = "</todo-edited>"
)

// TodoEditedBlock tells the model the user changed the todo list. It goes in
// front of the next user prompt.
func TodoEditedBlock(items []todo.Item) string {
	return todoEditedOpen + "The user edited the todo list. Current list:\n" + todo.Text(items) + "\n" + todoEditedClose + "\n\n"
}

// StripTodoEdited undoes TodoEditedBlock, leaving what the user typed. A
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
	undoOpen  = "<files-undone>"
	undoClose = "</files-undone>"
)

// UndoBlock tells the model the user reverted its file changes. It goes in
// front of the next user prompt.
func UndoBlock(paths []string) string {
	return undoOpen + "The user undid your file changes from your last turn that changed files. These files are back to how they were before it: " +
		strings.Join(paths, ", ") + ". Read them again before you change them." + undoClose + "\n\n"
}

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
