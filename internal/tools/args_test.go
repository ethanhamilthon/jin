package tools

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestWriteRequiresContent(t *testing.T) {
	path := writeTempFile(t, "keep")
	tests := []struct {
		name, args, wantErr string
		wantFile            string
	}{
		{"missing", `{"path":"P"}`, "missing required field: content", "keep"},
		{"null", `{"path":"P","content":null}`, "missing required field: content", "keep"},
		{"empty", `{"path":"P","content":""}`, "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.WriteFile(path, []byte("keep"), 0o644)
			_, err := (Write{}).Run(context.Background(), strings.ReplaceAll(tt.args, "P", path))
			if (err == nil) != (tt.wantErr == "") || (err != nil && err.Error() != tt.wantErr) {
				t.Fatalf("err=%v want %q", err, tt.wantErr)
			}
			data, _ := os.ReadFile(path)
			if string(data) != tt.wantFile {
				t.Fatalf("file=%q want %q", data, tt.wantFile)
			}
		})
	}
}

func TestEditRequiresNewString(t *testing.T) {
	path := writeTempFile(t, "abc def")
	tests := []struct {
		name, args, wantErr, wantFile string
	}{
		{"missing", `{"path":"P","old_string":"abc "}`, "missing required field: new_string", "abc def"},
		{"null", `{"path":"P","old_string":"abc ","new_string":null}`, "missing required field: new_string", "abc def"},
		{"empty deletes", `{"path":"P","old_string":"abc ","new_string":""}`, "", "def"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.WriteFile(path, []byte("abc def"), 0o644)
			_, err := (Edit{}).Run(context.Background(), strings.ReplaceAll(tt.args, "P", path))
			if (err == nil) != (tt.wantErr == "") || (err != nil && err.Error() != tt.wantErr) {
				t.Fatalf("err=%v want %q", err, tt.wantErr)
			}
			data, _ := os.ReadFile(path)
			if string(data) != tt.wantFile {
				t.Fatalf("file=%q want %q", data, tt.wantFile)
			}
		})
	}
}
