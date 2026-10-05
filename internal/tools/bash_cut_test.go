package tools

import (
	"context"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"
)

func TestBashLongOutputKeepsLogAndSaysWhere(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	out := runBash(context.Background(), "seq 1 20000", 10*time.Second, "", true)
	if !strings.HasPrefix(out, "1\n2\n") || !strings.Contains(out, "20000\n") {
		t.Fatalf("head or tail missing: %q ... %q", out[:20], out[len(out)-40:])
	}
	match := regexp.MustCompile(`The full output is in (\S+);`).FindStringSubmatch(out)
	if match == nil || !strings.Contains(out, "cut from the middle of") {
		t.Fatalf("no note: %q", out)
	}
	data, err := os.ReadFile(match[1])
	if err != nil || !strings.HasSuffix(string(data), "20000\n") {
		t.Fatalf("log not kept: %v", err)
	}
}
