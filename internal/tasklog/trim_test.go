package tasklog

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestTrimKeepsHeadAndTail(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.log")
	cases := []struct {
		name, content, want string
	}{
		{"small file stays", "short", "short"},
		{"big file is cut", "HEADxxxx" + strings.Repeat("-", 100) + "yyyyTAIL", "HEADxxxx" + cutNote + "yyyyTAIL"},
	}
	for _, c := range cases {
		if err := os.WriteFile(path, []byte(c.content), 0o600); err != nil {
			t.Fatal(err)
		}
		if err := trim(path, 32); err != nil {
			t.Fatal(err)
		}
		if got, _ := os.ReadFile(path); string(got) != c.want {
			t.Errorf("%s: got %q", c.name, got)
		}
	}
}

func TestTrimTwiceKeepsTheFirstHead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "x.log")
	_ = os.WriteFile(path, []byte("HEAD"+strings.Repeat("a", 60)), 0o600)
	_ = trim(path, 16)
	file, _ := os.OpenFile(path, os.O_WRONLY|os.O_APPEND, 0)
	file.WriteString(strings.Repeat("b", 60) + "LAST")
	file.Close()
	_ = trim(path, 16)
	got, _ := os.ReadFile(path)
	if !strings.HasPrefix(string(got), "HEAD") || !strings.HasSuffix(string(got), "LAST") || strings.Count(string(got), cutNote) != 1 {
		t.Errorf("got %q", got)
	}
}
