package core

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/provider"
	"jin/internal/tools"
)

func testPNG(w, h int) provider.Image {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h)))
	return provider.Image{MimeType: "image/png", Data: base64.StdEncoding.EncodeToString(buf.Bytes())}
}

func TestPictureLabelHasNameTypeAndSize(t *testing.T) {
	label := PictureLabel("/tmp/dir/shot.png", testPNG(30, 20))
	if !strings.HasPrefix(label, "[shot.png · image/png · ") || !strings.HasSuffix(label, " · 30×20]") {
		t.Fatalf("label = %q", label)
	}
}

func TestPictureLabelWithoutPathHasNoName(t *testing.T) {
	if label := PictureLabel("", testPNG(2, 2)); !strings.HasPrefix(label, "[image/png · ") {
		t.Fatalf("label = %q", label)
	}
}

func TestRunToolsShowsPicturesWithTheirFileNames(t *testing.T) {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 4, 4)))
	dir := t.TempDir()
	var calls = []provider.ToolCall{}
	for i, name := range []string{"one.png", "two.png"} {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, buf.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		calls = append(calls, readCall(t, string(rune('a'+i)), path))
	}
	agent := NewAgent(nil, "sys", tools.NewRegistry(tools.NewRead()))
	var history []provider.Message
	updates := make(chan Update, 16)
	if err := agent.runTools(t.Context(), t.Context(), Request{}, calls, &history, updates); err != nil {
		t.Fatal(err)
	}
	close(updates)
	var labels []string
	for u := range updates {
		if u.Image != nil {
			labels = append(labels, u.Text)
		}
	}
	if len(labels) != 2 || !strings.Contains(labels[0], "one.png") || !strings.Contains(labels[1], "two.png") {
		t.Fatalf("labels = %q", labels)
	}
}
