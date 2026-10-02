package ui

import (
	"fmt"
	"os"

	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
	"jin/internal/todo"
)

func isTodoKey(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyCtrlT {
		return true
	}
	return ev.Key() == tcell.KeyRune && ev.Modifiers()&tcell.ModCtrl != 0 && (ev.Str() == "t" || ev.Str() == "T")
}

// editTodos opens the todo list in the editor. Nothing is saved when the
// text is unchanged or does not parse; a broken file is kept so the edit is
// not lost.
func (a *app) editTodos(s *chatSession) {
	if !s.persisted {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "No todo list yet"})
		return
	}
	if err := a.runTodoEdit(s); err != nil {
		s.closeOpenEntry()
		s.appendEntry(chatEntry{kind: core.UpdateError, text: err.Error()})
	}
}

func (a *app) runTodoEdit(s *chatSession) error {
	items, err := s.store.LoadTodos(s.id)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "No todo list yet"})
		return nil
	}
	file, err := os.CreateTemp("", "jin-todo-*.md")
	if err != nil {
		return err
	}
	before := todo.Format(items)
	_, writeErr := file.WriteString(before)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		os.Remove(file.Name())
		return fmt.Errorf("cannot write the todo file: %v", firstErr(writeErr, closeErr))
	}
	return a.editFile(file.Name(), func() error {
		data, err := os.ReadFile(file.Name())
		if err != nil {
			return err
		}
		edited, err := todo.Parse(string(data))
		if err != nil {
			return fmt.Errorf("todo not saved, %v. Your edit is kept in %s", err, file.Name())
		}
		os.Remove(file.Name())
		if todo.Equal(edited, items) {
			return nil
		}
		if err := s.store.SaveTodos(s.id, edited); err != nil {
			return err
		}
		if err := s.store.MarkTodosEdited(s.id); err != nil {
			return err
		}
		s.setTodos(edited)
		return nil
	})
}

func firstErr(errs ...error) error {
	for _, err := range errs {
		if err != nil {
			return err
		}
	}
	return nil
}
