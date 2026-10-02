package hooks

import (
	"os"
	"slices"
	"testing"
)

func TestCreateListDeleteAndRender(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	if names, err := List(); err != nil || names != nil {
		t.Fatalf("empty dir: %v, %v", names, err)
	}
	for _, name := range []string{"b-style", "a-tone", "empty"} {
		if _, err := Create(name); err != nil {
			t.Fatal(err)
		}
	}
	write(t, "b-style", "Be brief.\n")
	write(t, "a-tone", "Be kind.")
	if names, _ := List(); !slices.Equal(names, []string{"a-tone", "b-style", "empty"}) {
		t.Errorf("names = %v", names)
	}
	got := Render(nil)
	if got != "Be kind.\n\nBe brief." {
		t.Errorf("Render = %q", got)
	}
	if got := Render([]string{"a-tone"}); got != "Be brief." {
		t.Errorf("Render with a-tone off = %q", got)
	}
	if err := Delete("a-tone"); err != nil {
		t.Fatal(err)
	}
	if names, _ := List(); slices.Contains(names, "a-tone") {
		t.Errorf("a-tone still listed: %v", names)
	}
}

// docs and async used to be built-in hooks. They are not any more, so the
// names are free for the user.
func TestFormerSystemNamesAreOrdinaryHooks(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for _, name := range []string{"docs", "async"} {
		if _, err := Create(name); err != nil {
			t.Fatalf("Create(%q): %v", name, err)
		}
		write(t, name, "my "+name)
	}
	if names, _ := List(); !slices.Equal(names, []string{"async", "docs"}) {
		t.Errorf("names = %v", names)
	}
	if got := Render(nil); got != "my async\n\nmy docs" {
		t.Errorf("Render = %q", got)
	}
}

func TestBadNames(t *testing.T) {
	for _, name := range []string{"", ".hidden", "a/b", "../x", "a b"} {
		if _, err := Path(name); err == nil {
			t.Errorf("Path(%q) should fail", name)
		}
	}
}

func TestToggleAndForget(t *testing.T) {
	off := Toggle(nil, "a")
	if !slices.Equal(off, []string{"a"}) || len(Toggle(off, "a")) != 0 {
		t.Fatalf("Toggle mismatch: %v", off)
	}
	if got := Forget([]string{"a", "b"}, "a"); !slices.Equal(got, []string{"b"}) {
		t.Errorf("Forget = %v", got)
	}
}

func write(t *testing.T, name, body string) {
	t.Helper()
	path, err := Path(name)
	if err != nil || os.WriteFile(path, []byte(body), 0o600) != nil {
		t.Fatal(err)
	}
}
