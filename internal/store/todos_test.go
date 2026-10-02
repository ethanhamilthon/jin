package store

import (
	"testing"

	"jin/internal/todo"
)

func TestTodosRoundTrip(t *testing.T) {
	db, second := openTwo(t)
	long := []todo.Item{{Text: "a", Status: todo.Pending}, {Text: "b", Status: todo.InProgress}, {Text: "c", Status: todo.Done}}
	if err := db.SaveTodos("s1", long); err != nil {
		t.Fatal(err)
	}
	short := []todo.Item{{Text: "z", Status: todo.Done}}
	if err := db.SaveTodos("s1", short); err != nil {
		t.Fatal(err)
	}
	got, err := second.LoadTodos("s1")
	if err != nil || !todo.Equal(got, short) {
		t.Fatalf("got %v, %v", got, err)
	}
	if other, _ := second.LoadTodos("s2"); len(other) != 0 {
		t.Fatalf("other session: %v", other)
	}
}

func TestTodosEditedFlag(t *testing.T) {
	db, _ := openTwo(t)
	if ok, _ := db.TakeTodosEdited("s1"); ok {
		t.Fatal("flag set by default")
	}
	if err := db.MarkTodosEdited("s1"); err != nil {
		t.Fatal(err)
	}
	if ok, _ := db.TodosEdited("s1"); !ok {
		t.Fatal("peek lost flag")
	}
	if ok, _ := db.TakeTodosEdited("s1"); !ok {
		t.Fatal("flag not taken")
	}
	if ok, _ := db.TakeTodosEdited("s1"); ok {
		t.Fatal("flag not cleared")
	}
}
