package ui

import (
	"jin/internal/core"
	"jin/internal/prompts"
	"jin/internal/session"
	"jin/internal/todo"
	"strings"
)

func (s *chatSession) send(text string) { s.sendFiles(text, text, "", nil) }

// sendFiles shows text in the chat and sends clean plus the attached-files
// block to the model, with the bodies of the named prompts.
func (s *chatSession) sendFiles(text, clean, block string, names []string) {
	s.closeOpenEntry()
	prompt := prompts.ExpandNames(clean, names, s.promptBodies)
	if block != "" {
		prompt += "\n\n" + block
	}
	note := s.undoNote + s.todoNote()
	s.undoNote = ""
	request := core.Request{Prompt: note + prompt, Model: s.model, Effort: s.effort, Window: s.window(), NoVision: s.noVision()}
	if strings.TrimSpace(request.Prompt) != "" {
		// A command that runs now moves to the background, not in your way.
		request.Interactive = true
		s.agent.Expect()
	}
	s.pending = append(s.pending, request)
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: text})
	s.scroll = 0
	s.touch(text)
}

// todoNote tells the model about a todo list the user edited, once.
func (s *chatSession) todoNote() string {
	if !s.persisted {
		return ""
	}
	edited, err := s.store.TakeTodosEdited(s.id)
	if err != nil || !edited {
		return ""
	}
	items, err := s.store.LoadTodos(s.id)
	if err != nil {
		return ""
	}
	return core.TodoEditedBlock(items)
}

// setTodos shows a new list. A finished list leaves the pin and goes into
// the timeline once.
func (s *chatSession) setTodos(items []todo.Item) {
	s.todos = items
	if todo.AllDone(items) {
		s.appendEntry(chatEntry{kind: core.UpdateTodo, text: session.TodoText(items)})
	}
}

// pinnedTodos is the list shown above the input, if any.
func (s *chatSession) pinnedTodos() []todo.Item {
	if len(s.todos) == 0 || todo.AllDone(s.todos) {
		return nil
	}
	return s.todos
}
