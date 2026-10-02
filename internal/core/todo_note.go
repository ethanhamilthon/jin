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
