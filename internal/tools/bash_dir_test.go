package tools

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBashRunsInDir(t *testing.T) {
	base := t.TempDir()
	if err := os.MkdirAll(filepath.Join(base, "sub", "deep"), 0o755); err != nil {
		t.Fatal(err)
	}
	bash := Bash{dir: base}
	out, err := bash.Run(context.Background(), `{"command":"pwd","dir":"sub/deep"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(strings.TrimSpace(out), filepath.Join("sub", "deep")) {
		t.Fatalf("expected the command to start in sub/deep, got %q", out)
	}
	out, _ = bash.Run(context.Background(), `{"command":"pwd"}`)
	if strings.TrimSpace(out) != base {
		t.Fatalf("expected the next call to start in the base again, got %q", out)
	}
	out, _ = bash.Run(context.Background(), `{"command":"pwd","dir":"`+os.TempDir()+`"}`)
	if !strings.Contains(out, os.TempDir()) {
		t.Fatalf("expected an absolute dir to be used as is, got %q", out)
	}
}

func TestBashRejectsMissingDir(t *testing.T) {
	bash := Bash{dir: t.TempDir()}
	if _, err := bash.Run(context.Background(), `{"command":"pwd","dir":"nope"}`); err == nil || !strings.Contains(err.Error(), "does not exist") {
		t.Fatalf("expected a missing dir error, got %v", err)
	}
	if _, err := bash.Run(context.Background(), `{"command":"pwd","dir":"`+os.Args[0]+`"}`); err == nil || !strings.Contains(err.Error(), "not a directory") {
		t.Fatalf("expected a not-a-directory error, got %v", err)
	}
}

func TestBashSummaryNamesDir(t *testing.T) {
	got, ok := Bash{}.Summary(`{"command":"ls","dir":"web"}`)
	if !ok || got != "ls (120s) in web" {
		t.Fatalf("got %q ok=%v", got, ok)
	}
}
