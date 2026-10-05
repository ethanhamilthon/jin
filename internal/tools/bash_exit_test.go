package tools

import (
	"context"
	"strings"
	"testing"
)

func TestBashReportsNonZeroExitNeutrally(t *testing.T) {
	out, err := Bash{}.Run(context.Background(), `{"command":"echo searching; grep -q needle /dev/null"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.HasSuffix(out, "[exit code: 1]") || strings.Contains(out, "failed") {
		t.Fatalf("expected a neutral exit code, got %q", out)
	}
}

func TestBashSuccessHasNoExitNote(t *testing.T) {
	out, err := Bash{}.Run(context.Background(), `{"command":"echo ok"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if strings.Contains(out, "exit code") {
		t.Fatalf("expected no exit note on success, got %q", out)
	}
}
