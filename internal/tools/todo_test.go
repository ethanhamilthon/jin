package tools

import (
	"context"
	"errors"
	"strings"
	"testing"

	"jin/internal/todo"
)

type fakeTodos struct {
	items  []todo.Item
	edited bool
}

func (f *fakeTodos) Load() ([]todo.Item, error) { return f.items, nil }
func (f *fakeTodos) Save(i []todo.Item) error   { f.items = i; return nil }
func (f *fakeTodos) TakeEdited() (bool, error) {
	e := f.edited
	f.edited = false
	return e, nil
}

func TestTodoReplacesAndReads(t *testing.T) {
	store := &fakeTodos{}
	tool := NewTodo(store)
	var sunk []todo.Item
	ctx := WithTodoSink(context.Background(), func(i []todo.Item) { sunk = i })
	out, err := tool.Run(ctx, `{"items":[{"text":"a\nb","status":"in_progress"},{"text":"c","status":"in_progress"}]}`)
	if err != nil || len(store.items) != 2 || store.items[0].Text != "a b" || len(sunk) != 2 {
		t.Fatalf("out=%q err=%v items=%v", out, err, store.items)
	}
	out, err = tool.Run(ctx, `{}`)
	if err != nil || !strings.Contains(out, "[~] a b") || len(store.items) != 2 {
		t.Fatalf("read: %q %v", out, err)
	}
	if _, err := tool.Run(ctx, `{"items":[{"text":"x","status":"bogus"}]}`); err == nil {
		t.Fatal("bad status accepted")
	}
}

func TestTodoRefusesUpdateAfterUserEdit(t *testing.T) {
	store := &fakeTodos{items: []todo.Item{{Text: "kept", Status: todo.Pending}}, edited: true}
	tool := NewTodo(store)
	out, _ := tool.Run(context.Background(), `{"items":[{"text":"stale","status":"done"}]}`)
	if !strings.Contains(out, "not applied") || store.items[0].Text != "kept" {
		t.Fatalf("update applied: %q %v", out, store.items)
	}
	if _, err := tool.Run(context.Background(), `{"items":[{"text":"fresh","status":"done"}]}`); err != nil || store.items[0].Text != "fresh" {
		t.Fatalf("second update refused: %v %v", err, store.items)
	}
}

func TestAskFormatsAnswersAndHonorsCancel(t *testing.T) {
	ctx := WithAsker(context.Background(), func(_ context.Context, q []Question) ([]string, error) {
		return []string{"yes", "blue"}, nil
	})
	out, err := NewAsk().Run(ctx, `{"questions":[{"question":"Ok?","options":["yes","no"]},{"question":"Color?"}]}`)
	if err != nil || out != "Ok? → yes\nColor? → blue" {
		t.Fatalf("out=%q err=%v", out, err)
	}
	cancelled := WithAsker(context.Background(), func(context.Context, []Question) ([]string, error) {
		return nil, context.Canceled
	})
	if _, err := NewAsk().Run(cancelled, `{"questions":[{"question":"x"}]}`); !errors.Is(err, context.Canceled) {
		t.Fatalf("err=%v", err)
	}
	if _, ok := NewAsk().Summary(`{"questions":[]}`); ok {
		t.Fatal("empty questions accepted")
	}
}

func TestEmptyRegistryHasNoSchema(t *testing.T) {
	if NewRegistry().SchemaJSON() != nil {
		t.Fatal("schema not nil")
	}
	if got := Build([]string{"read", "todo"}, nil).Names(); len(got) != 1 {
		t.Fatalf("todo without store built: %v", got)
	}
}
