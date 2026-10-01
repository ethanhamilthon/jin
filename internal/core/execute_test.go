package core

import (
	"bytes"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

func readCall(t *testing.T, id, path string) provider.ToolCall {
	t.Helper()
	args, _ := json.Marshal(map[string]string{"path": path})
	call := provider.ToolCall{ID: id, Type: "function"}
	call.Function.Name, call.Function.Arguments = "read", string(args)
	return call
}

func TestRunToolsAttachesPicturesAfterTheLastToolMessage(t *testing.T) {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4)))
	path := filepath.Join(t.TempDir(), "shot.png")
	if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
		t.Fatal(err)
	}
	agent := NewAgent(nil, "sys", tools.NewRegistry(tools.NewRead()))
	calls := []provider.ToolCall{readCall(t, "a", path), readCall(t, "b", path)}
	var history []provider.Message
	updates := make(chan Update, 16)
	if err := agent.runTools(t.Context(), t.Context(), Request{}, calls, &history, updates); err != nil {
		t.Fatal(err)
	}
	if len(history) != 3 || history[0].Role != "tool" || history[1].Role != "tool" {
		t.Fatalf("history = %+v", history)
	}
	if last := history[2]; last.Role != "user" || len(last.Images) != 2 || ImageLabels(last) == nil {
		t.Fatalf("attachment = %+v", last)
	}
}

func TestRunToolsOmitsPicturesForTextOnlyModel(t *testing.T) {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4)))
	path := filepath.Join(t.TempDir(), "shot.png")
	_ = os.WriteFile(path, buf.Bytes(), 0o600)
	agent := NewAgent(nil, "sys", tools.NewRegistry(tools.NewRead()))
	var history []provider.Message
	updates := make(chan Update, 16)
	if err := agent.runTools(t.Context(), t.Context(), Request{NoVision: true}, []provider.ToolCall{readCall(t, "a", path)}, &history, updates); err != nil {
		t.Fatal(err)
	}
	if len(history) != 1 || !strings.Contains(history[0].Content, "does not support images") {
		t.Fatalf("history = %+v", history)
	}
}
