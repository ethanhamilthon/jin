package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"jin/internal/todo"
)

const todoSchema = `{"type":"function","function":{"name":"todo","description":"Keep a todo list for the current session. Send the whole list on every call; a call replaces the list. Call without items to read the current list. Several items may be in_progress at once.","parameters":{"type":"object","properties":{"items":{"type":"array","description":"The complete list. Omit to only read it.","items":{"type":"object","properties":{"text":{"type":"string","description":"One line describing the task"},"status":{"type":"string","enum":["pending","in_progress","done"]}},"required":["text","status"],"additionalProperties":false}}},"additionalProperties":false}}}`

// TodoStore keeps the todo list of one session.
type TodoStore interface {
	Load() ([]todo.Item, error)
	Save([]todo.Item) error
	// TakeEdited reports whether the user edited the list since the model
	// last saw it, and clears the flag.
	TakeEdited() (bool, error)
}

type todoSinkKey struct{}

// WithTodoSink makes the todo tool report every changed list to fn.
func WithTodoSink(ctx context.Context, fn func([]todo.Item)) context.Context {
	return context.WithValue(ctx, todoSinkKey{}, fn)
}

type Todo struct{ store TodoStore }

func NewTodo(store TodoStore) Todo { return Todo{store: store} }

func (Todo) Name() string { return "todo" }

func (Todo) Schema() json.RawMessage { return json.RawMessage(todoSchema) }

type todoArgs struct {
	Items *[]todo.Item `json:"items"`
}

func (Todo) Summary(argumentsJSON string) (string, bool) {
	var args todoArgs
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil {
		return "", false
	}
	if args.Items == nil {
		return "read list", true
	}
	done, total := todo.Counts(*args.Items)
	return strconv.Itoa(done) + "/" + strconv.Itoa(total) + " done", true
}

func (t Todo) Run(ctx context.Context, argumentsJSON string) (string, error) {
	var args todoArgs
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil {
		return "", errors.New("invalid todo tool arguments")
	}
	current, err := t.store.Load()
	if err != nil {
		return "", err
	}
	edited, err := t.store.TakeEdited()
	if err != nil {
		return "", err
	}
	if args.Items == nil {
		return todo.Text(current), nil
	}
	if edited {
		return "The user edited the todo list; your update was not applied. Current list:\n" + todo.Text(current), nil
	}
	items, err := todo.Validate(*args.Items)
	if err != nil {
		return "", err
	}
	if err := t.store.Save(items); err != nil {
		return "", err
	}
	if sink, ok := ctx.Value(todoSinkKey{}).(func([]todo.Item)); ok {
		sink(items)
	}
	return "Todo list saved:\n" + todo.Text(items), nil
}

// MemoryTodos is a TodoStore that lives only as long as the process, for
// runs that do not save the session.
type MemoryTodos struct{ items []todo.Item }

func (m *MemoryTodos) Load() ([]todo.Item, error) { return m.items, nil }
func (m *MemoryTodos) Save(items []todo.Item) error {
	m.items = items
	return nil
}
func (m *MemoryTodos) TakeEdited() (bool, error) { return false, nil }
