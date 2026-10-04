package tools

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func runRead(t *testing.T, args map[string]any) (string, error) {
	t.Helper()
	raw, _ := json.Marshal(args)
	return (Read{}).Run(context.Background(), string(raw))
}

func TestReadRanges(t *testing.T) {
	path := writeTempFile(t, "one\ntwo\nthree\nfour\n")
	tests := []struct {
		name string
		args map[string]any
		want string
	}{
		{"whole file", map[string]any{"path": path}, "1\tone\n2\ttwo\n3\tthree\n4\tfour\n"},
		{"offset", map[string]any{"path": path, "offset": 3}, "3\tthree\n4\tfour\n"},
		{"offset and limit", map[string]any{"path": path, "offset": 2, "limit": 2}, "2\ttwo\n3\tthree\n"},
		{"past the end", map[string]any{"path": path, "offset": 9}, "[file has 4 lines; offset 9 is past the end]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := runRead(t, tt.args)
			if err != nil || got != tt.want {
				t.Fatalf("got %q err=%v, want %q", got, err, tt.want)
			}
		})
	}
}

func TestReadCapsLongLine(t *testing.T) {
	path := writeTempFile(t, strings.Repeat("é", 3_000_000)+"\nnext\n")
	got, err := runRead(t, map[string]any{"path": path})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(got) > maxReadOutput {
		t.Fatalf("output is %d bytes", len(got))
	}
	if !strings.Contains(got, "... [line truncated, 3000000 chars]\n2\tnext\n") {
		t.Fatalf("missing truncation note: %.80q", got[len(got)-80:])
	}
}

func TestReadStopsAtBudgetBeforeAppending(t *testing.T) {
	var lines []string
	for i := 0; i < 100; i++ {
		lines = append(lines, strings.Repeat("x", 1000))
	}
	path := writeTempFile(t, strings.Join(lines, "\n"))
	got, err := runRead(t, map[string]any{"path": path})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	note := "[truncated at line 32; continue with offset=33]"
	if !strings.HasSuffix(got, note) || len(got) > maxReadOutput+len(note) {
		t.Fatalf("len=%d suffix=%q", len(got), got[len(got)-60:])
	}
	if strings.Contains(got, "\n33\t") {
		t.Fatal("line 33 must not be included")
	}
	rest, _ := runRead(t, map[string]any{"path": path, "offset": 33, "limit": 1})
	if rest != "33\t"+lines[0]+"\n" {
		t.Fatalf("continue offset did not resume at line 33")
	}
}

func TestReadRefusesBinary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "blob.bin")
	os.WriteFile(path, append([]byte("abc"), 0, 1, 2), 0o644)
	_, err := runRead(t, map[string]any{"path": path})
	if err == nil || !strings.Contains(err.Error(), "binary file") {
		t.Fatalf("expected binary refusal, got %v", err)
	}
}

func TestReadNulAfterSniffWindowIsText(t *testing.T) {
	path := writeTempFile(t, strings.Repeat("a", binarySniff+10)+"\x00\n")
	if _, err := runRead(t, map[string]any{"path": path}); err != nil {
		t.Fatalf("Run: %v", err)
	}
}

func TestReadRejectsDirectory(t *testing.T) {
	if _, err := runRead(t, map[string]any{"path": t.TempDir()}); err == nil {
		t.Fatal("expected error for directory")
	}
}

func TestReadLineNumbersSurviveLongLines(t *testing.T) {
	path := writeTempFile(t, strings.Repeat("y", 70_000)+"\nsecond\n")
	got, _ := runRead(t, map[string]any{"path": path, "offset": 2})
	if got != "2\tsecond\n" {
		t.Fatalf("got %q", got)
	}
	_ = strconv.Itoa
}

func TestReadStillAttachesPictures(t *testing.T) {
	path := filepath.Join(t.TempDir(), "p.png")
	os.WriteFile(path, pngBytes(t, 10, 10), 0o644)
	raw, _ := json.Marshal(map[string]any{"path": path})
	_, images, err := (Read{}).RunImages(context.Background(), string(raw))
	if err != nil || len(images) != 1 {
		t.Fatalf("images=%d err=%v", len(images), err)
	}
}
