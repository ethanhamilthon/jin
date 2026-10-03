package datadir

import (
	"os"
	"path/filepath"
	"testing"
)

func seed(t *testing.T, dir, marker string) {
	t.Helper()
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(dir, "jin.db"), []byte(marker), 0o600)
}

func marker(dir string) string {
	data, _ := os.ReadFile(filepath.Join(dir, "jin.db"))
	return string(data)
}

func TestResetAndSwap(t *testing.T) {
	root := t.TempDir()
	current, backup := filepath.Join(root, ".jin"), filepath.Join(root, "old-jin")
	seed(t, current, "first")
	if err := Reset(current, filepath.Join(current, "inner")); err == nil {
		t.Fatal("dest inside the data dir must fail")
	}
	if err := Reset(current, backup); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(current); !os.IsNotExist(err) || marker(backup) != "first" {
		t.Fatal("reset did not move the folder")
	}
	seed(t, current, "second")
	if err := Reset(current, backup); err == nil {
		t.Fatal("existing dest must fail")
	}
	if err := Swap(current, backup); err != nil {
		t.Fatal(err)
	}
	if marker(current) != "first" || marker(backup) != "second" {
		t.Fatalf("swap: %q %q", marker(current), marker(backup))
	}
	empty := filepath.Join(root, "empty")
	_ = os.Mkdir(empty, 0o700)
	if err := Swap(current, empty); err == nil {
		t.Fatal("folder without jin.db must fail")
	}
}

func TestExpand(t *testing.T) {
	t.Setenv("HOME", "/home/u")
	if got, _ := Expand(" ~/backup "); got != "/home/u/backup" {
		t.Fatalf("expand = %q", got)
	}
	if _, err := Expand(""); err == nil {
		t.Fatal("empty path must fail")
	}
}
