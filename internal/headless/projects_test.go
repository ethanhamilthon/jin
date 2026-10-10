package headless

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectsAddRenameArchiveRestore(t *testing.T) {
	h := newHarness(t, nil)
	h.dir = t.TempDir()
	app := filepath.Join(h.dir, "app")
	if err := os.Mkdir(app, 0o755); err != nil {
		t.Fatal(err)
	}
	want, err := filepath.EvalSymlinks(app)
	if err != nil {
		t.Fatal(err)
	}
	row := func(state string) string { return want + "\tapp\t" + state + "\n" }

	if code := h.run(t, "projects", "add", "app"); code != 0 || h.out.String() != row("active") {
		t.Fatalf("add: %d %q %q", code, h.out.String(), h.errOut.String())
	}
	if code := h.run(t, "projects", "list"); code != 0 || h.out.String() != row("active") {
		t.Fatalf("list: %d %q", code, h.out.String())
	}
	if code := h.run(t, "projects", "rename", "app", "Shop"); code != 0 || h.out.String() != want+"\tShop\tactive\n" {
		t.Fatalf("rename: %d %q %q", code, h.out.String(), h.errOut.String())
	}
	if code := h.run(t, "projects", "archive", app); code != 0 || h.out.String() != want+"\tShop\tarchived\n" {
		t.Fatalf("archive: %d %q %q", code, h.out.String(), h.errOut.String())
	}
	if code := h.run(t, "projects", "list"); code != 0 || h.out.String() != "" {
		t.Fatalf("list hides archived: %d %q", code, h.out.String())
	}
	if code := h.run(t, "projects", "list", "--all"); code != 0 || h.out.String() != want+"\tShop\tarchived\n" {
		t.Fatalf("list --all: %d %q", code, h.out.String())
	}
	if code := h.run(t, "projects", "restore", "app"); code != 0 || h.out.String() != want+"\tShop\tactive\n" {
		t.Fatalf("restore: %d %q %q", code, h.out.String(), h.errOut.String())
	}
	h.run(t, "projects", "archive", "app")
	if code := h.run(t, "projects", "add", "app"); code != 0 || h.out.String() != want+"\tShop\tactive\n" {
		t.Fatalf("add restores: %d %q", code, h.out.String())
	}
}

func TestProjectsListJSON(t *testing.T) {
	h := newHarness(t, nil)
	if code := h.run(t, "projects", "list", "--format", "json"); code != 0 || h.out.String() != "[]\n" {
		t.Fatalf("empty: %d %q", code, h.out.String())
	}
	h.dir = t.TempDir()
	if code := h.run(t, "projects", "add", h.dir); code != 0 {
		t.Fatalf("add: %d %q", code, h.errOut.String())
	}
	if code := h.run(t, "projects", "list", "--format", "json"); code != 0 ||
		!strings.Contains(h.out.String(), `"archived":false`) || !strings.Contains(h.out.String(), `"path":`) {
		t.Fatalf("json: %d %q", code, h.out.String())
	}
}

func TestProjectsErrors(t *testing.T) {
	h := newHarness(t, nil)
	h.dir = t.TempDir()
	cases := map[string][]string{
		"no subcommand":     {"projects"},
		"unknown":           {"projects", "remove", "x"},
		"add missing dir":   {"projects", "add", "nope"},
		"rename unknown":    {"projects", "rename", h.dir, "Name"},
		"archive unknown":   {"projects", "archive", h.dir},
		"rename no name":    {"projects", "rename", h.dir},
		"archive two paths": {"projects", "archive", h.dir, "x"},
		"bad format":        {"projects", "list", "--format", "xml"},
		"list extra word":   {"projects", "list", "extra"},
	}
	for name, args := range cases {
		h.errOut.Reset()
		if code := h.run(t, args...); code != 1 || !strings.Contains(h.errOut.String(), "jin") {
			t.Errorf("%s: code %d, stderr %q", name, code, h.errOut.String())
		}
		if h.out.Len() != 0 {
			t.Errorf("%s: stdout %q", name, h.out.String())
		}
	}
}

func TestProjectsHandles(t *testing.T) {
	if !Handles([]string{"projects", "list"}) {
		t.Fatal("projects is not a headless command")
	}
}
