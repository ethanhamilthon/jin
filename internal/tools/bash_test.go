package tools

import (
	"context"
	"strings"
	"testing"
)

func TestBashRunRejectsZeroTimeout(t *testing.T) {
	_, err := Bash{}.Run(context.Background(), `{"command":"echo hi","timeout":0}`)
	if err == nil {
		t.Fatal("expected an error for timeout: 0")
	}
}

func TestBashRunRejectsNegativeTimeout(t *testing.T) {
	_, err := Bash{}.Run(context.Background(), `{"command":"echo hi","timeout":-5}`)
	if err == nil {
		t.Fatal("expected an error for a negative timeout")
	}
}

func TestBashSummaryRejectsZeroTimeout(t *testing.T) {
	if _, ok := (Bash{}).Summary(`{"command":"echo hi","timeout":0}`); ok {
		t.Fatal("expected Summary to reject timeout: 0 before the tool ever runs")
	}
}

func TestBashSummaryAppendsTimeout(t *testing.T) {
	got, ok := (Bash{}).Summary(`{"command":"echo hi","timeout":5}`)
	if !ok || got != "echo hi (5s)" {
		t.Fatalf("expected explicit timeout appended, got %q ok=%v", got, ok)
	}
	got, ok = (Bash{}).Summary(`{"command":"echo hi"}`)
	if !ok || got != "echo hi (120s)" {
		t.Fatalf("expected default timeout appended, got %q ok=%v", got, ok)
	}
}

func TestBashRunUsesDefaultTimeoutWhenOmitted(t *testing.T) {
	out, err := Bash{}.Run(context.Background(), `{"command":"echo hi"}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "hi") {
		t.Fatalf("expected command output, got %q", out)
	}
}

func TestBashRunHonorsExplicitTimeout(t *testing.T) {
	out, err := Bash{}.Run(context.Background(), `{"command":"sleep 2","timeout":1}`)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if !strings.Contains(out, "[command timed out]") {
		t.Fatalf("expected the 1s timeout to fire before the 2s sleep finished, got %q", out)
	}
}
