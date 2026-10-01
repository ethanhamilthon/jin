package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pngBytes(t *testing.T, w, h int) []byte {
	t.Helper()
	var buf bytes.Buffer
	if err := png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, w, h))); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestLoadImageKeepsSmallPictureAsIs(t *testing.T) {
	data := pngBytes(t, 10, 10)
	got, ok := loadImage(data)
	if !ok || got.MimeType != "image/png" || got.Data != base64.StdEncoding.EncodeToString(data) {
		t.Fatalf("got %+v, ok = %v", got, ok)
	}
}

func TestLoadImageShrinksLargePicture(t *testing.T) {
	got, ok := loadImage(pngBytes(t, 3000, 1500))
	if !ok {
		t.Fatal("not loaded")
	}
	raw, _ := base64.StdEncoding.DecodeString(got.Data)
	config, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || config.Width != maxImageSide || config.Height != maxImageSide/2 {
		t.Fatalf("size = %dx%d, err = %v", config.Width, config.Height, err)
	}
}

func TestLoadImageRejectsText(t *testing.T) {
	if _, ok := loadImage([]byte("package main\n")); ok {
		t.Fatal("text was taken for an image")
	}
}

func TestReadReturnsPictureAndKeepsText(t *testing.T) {
	dir := t.TempDir()
	picture := filepath.Join(dir, "shot.png")
	text := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(picture, pngBytes(t, 4, 4), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(text, []byte("hello\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	args := func(path string) string {
		out, _ := json.Marshal(map[string]string{"path": path})
		return string(out)
	}
	note, images, err := NewRead().RunImages(context.Background(), args(picture))
	if err != nil || len(images) != 1 || !strings.Contains(note, "image/png") {
		t.Fatalf("picture: note = %q, images = %d, err = %v", note, len(images), err)
	}
	note, images, err = NewRead().RunImages(context.Background(), args(text))
	if err != nil || len(images) != 0 || !strings.Contains(note, "hello") {
		t.Fatalf("text: note = %q, images = %d, err = %v", note, len(images), err)
	}
}
