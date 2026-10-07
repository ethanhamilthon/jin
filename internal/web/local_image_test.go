package web

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadImageFileRefusesOtherFiles(t *testing.T) {
	dir := t.TempDir()
	png := "\x89PNG\r\n\x1a\n" + "0000000000000000"
	for name, body := range map[string]string{"a.png": png, "notes.txt": "hello"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := readImageFile(resolveLocal("a.png", dir)); err != nil {
		t.Fatal(err)
	}
	if _, err := readImageFile(resolveLocal("notes.txt", dir)); err == nil {
		t.Fatal("a text file was served")
	}
	if _, err := readImageFile(resolveLocal("missing.png", dir)); err == nil {
		t.Fatal("a missing file was served")
	}
}
