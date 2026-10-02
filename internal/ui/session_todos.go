package ui

import (
	"jin/internal/store"
	"jin/internal/todo"
)

// sessionTodos keeps the todo list of one session in the database.
type sessionTodos struct {
	db *store.DB
	id string
}

func (t sessionTodos) Load() ([]todo.Item, error) { return t.db.LoadTodos(t.id) }
func (t sessionTodos) Save(items []todo.Item) error {
	return t.db.SaveTodos(t.id, items)
}
func (t sessionTodos) TakeEdited() (bool, error) { return t.db.TakeTodosEdited(t.id) }
