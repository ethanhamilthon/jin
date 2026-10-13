package desktop

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadWindow(t *testing.T) {
	dir := t.TempDir()
	missing := filepath.Join(dir, "none.json")
	if got := loadWindow(missing); got != (windowState{defaultWidth, defaultHeight}) {
		t.Fatalf("missing = %+v", got)
	}
	file := filepath.Join(dir, "w.json")
	_ = os.WriteFile(file, []byte(`{"Width":900,"Height":700}`), 0o600)
	if got := loadWindow(file); got != (windowState{900, 700}) {
		t.Fatalf("saved = %+v", got)
	}
	_ = os.WriteFile(file, []byte(`{"Width":10,"Height":10}`), 0o600)
	if got := loadWindow(file); got != (windowState{defaultWidth, defaultHeight}) {
		t.Fatalf("tiny = %+v", got)
	}
	_ = os.WriteFile(file, []byte(`junk`), 0o600)
	if got := loadWindow(file); got != (windowState{defaultWidth, defaultHeight}) {
		t.Fatalf("junk = %+v", got)
	}
}
