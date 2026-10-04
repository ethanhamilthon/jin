package ui

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCustomThemeFile(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	t.Cleanup(func() { applyTheme(themes[0]) })
	dir, _ := themesDir()
	_ = os.MkdirAll(dir, 0o755)
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("mine.json", `{"name":"Mine","base":"Nord","colors":{"bg":"#101010","todo_panel":"#202020"}}`)
	write("broken.json", `{"colors":{"bgg":"#000000"}}`)
	list, errs := customThemes()
	if len(list) != 1 || len(errs) != 1 {
		t.Fatalf("list=%v errs=%v", list, errs)
	}
	mine := themeByName("Mine")
	nord := builtinTheme("Nord")
	if mine.bg != 0x101010 || mine.todo != 0x202020 || mine.fg != nord.fg {
		t.Fatalf("theme = %+v", mine)
	}
	applyTheme(mine)
	if colorTodoPanel != rgb(0x202020) || colorAskPanel == colorBG {
		t.Fatal("panel colors not applied")
	}
	if _, err := parseTheme([]byte(`{"colors":{"bg":"red"}}`), "x"); err == nil {
		t.Fatal("bad color must fail")
	}
}
