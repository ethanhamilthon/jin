package tools

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func runEditArgs(t *testing.T, path, old, new string) error {
	t.Helper()
	args, _ := json.Marshal(map[string]any{"path": path, "old_string": old, "new_string": new})
	_, err := Edit{}.Run(context.Background(), string(args))
	return err
}

func TestEditMissHints(t *testing.T) {
	tests := []struct {
		name, content, old, cause string
	}{
		{"crlf", "a\r\nb\r\nc\r\n", "b\nc", "line endings"},
		{"trailing", "a\nb  \nc\n", "b\nc", "trailing whitespace"},
		{"tabs", "a\n\tb\nc\n", "    b", "tabs versus spaces"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := writeTempFile(t, tt.content)
			err := runEditArgs(t, path, tt.old, "x")
			if err == nil || !strings.Contains(err.Error(), tt.cause) || !strings.Contains(err.Error(), "line 2") {
				t.Fatalf("error = %v", err)
			}
			if !strings.Contains(err.Error(), "line-number prefix") {
				t.Fatalf("no copy advice: %v", err)
			}
			if data, _ := os.ReadFile(path); string(data) != tt.content {
				t.Fatalf("file was changed: %q", data)
			}
		})
	}
}

func TestEditMissWithoutHint(t *testing.T) {
	path := writeTempFile(t, "b  \nb  \n")
	err := runEditArgs(t, path, "b\n", "x")
	if err == nil || strings.Contains(err.Error(), "cause") {
		t.Fatalf("two normalized matches must not hint: %v", err)
	}
	err = runEditArgs(t, writeTempFile(t, "hello\n"), "nothing", "x")
	if err == nil || strings.Contains(err.Error(), ";") {
		t.Fatalf("error = %v", err)
	}
}

func TestEditRejectsIdenticalStrings(t *testing.T) {
	err := runEditArgs(t, writeTempFile(t, "hello\n"), "hello", "hello")
	if err == nil || !strings.Contains(err.Error(), "identical") {
		t.Fatalf("error = %v", err)
	}
}
