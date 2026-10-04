package tools

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"errors"
	"hash/crc32"
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
	got, err := loadImage(data)
	if err != nil || got.MimeType != "image/png" || got.Data != base64.StdEncoding.EncodeToString(data) {
		t.Fatalf("got %+v, err = %v", got, err)
	}
}

func TestLoadImageShrinksLargePicture(t *testing.T) {
	got, err := loadImage(pngBytes(t, 3000, 1500))
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := base64.StdEncoding.DecodeString(got.Data)
	config, _, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || config.Width != maxImageSide || config.Height != maxImageSide/2 {
		t.Fatalf("size = %dx%d, err = %v", config.Width, config.Height, err)
	}
}

func TestLoadImageRejectsText(t *testing.T) {
	if _, err := loadImage([]byte("package main\n")); !errors.Is(err, errNotImage) {
		t.Fatalf("text was taken for an image: %v", err)
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

func hugePNGHeader(w, h uint32) []byte {
	ihdr := make([]byte, 13)
	binary.BigEndian.PutUint32(ihdr[0:], w)
	binary.BigEndian.PutUint32(ihdr[4:], h)
	ihdr[8], ihdr[9] = 8, 6
	chunk := append([]byte("IHDR"), ihdr...)
	out := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\x0d")
	out = append(out, chunk...)
	return binary.BigEndian.AppendUint32(out, crc32.ChecksumIEEE(chunk))
}

func TestLoadImageRefusesTooManyPixelsBeforeDecoding(t *testing.T) {
	_, err := loadImage(hugePNGHeader(10000, 10000))
	if err == nil || errors.Is(err, errNotImage) || !strings.Contains(err.Error(), "40 megapixels") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadNamesImageErrors(t *testing.T) {
	dir := t.TempDir()
	read := func(name string, data []byte) error {
		path := filepath.Join(dir, name)
		if err := os.WriteFile(path, data, 0o600); err != nil {
			t.Fatal(err)
		}
		args, _ := json.Marshal(map[string]string{"path": path})
		_, _, err := NewRead().RunImages(context.Background(), string(args))
		return err
	}
	if err := read("huge.png", hugePNGHeader(10000, 10000)); err == nil || !strings.Contains(err.Error(), "40 megapixels") {
		t.Fatalf("huge: %v", err)
	}
	if err := read("cut.png", pngBytes(t, 3000, 1500)[:200]); err == nil || strings.Contains(err.Error(), "binary") {
		t.Fatalf("damaged picture must not be a binary refusal: %v", err)
	}
	big := append(pngBytes(t, 4, 4), make([]byte, maxImageFile)...)
	if err := read("big.png", big); err == nil || !strings.Contains(err.Error(), "limit is 20 MB") {
		t.Fatalf("big: %v", err)
	}
}

func TestLoadImageRefusesSmallHeaderOnlyPNG(t *testing.T) {
	_, err := loadImage(hugePNGHeader(10, 10))
	if err == nil || !strings.Contains(err.Error(), "damaged") {
		t.Fatalf("err = %v", err)
	}
}

func TestReadDamagedHeaderIsNotBinary(t *testing.T) {
	path := filepath.Join(t.TempDir(), "cut.png")
	if err := os.WriteFile(path, pngBytes(t, 4, 4)[:20], 0o600); err != nil {
		t.Fatal(err)
	}
	args, _ := json.Marshal(map[string]string{"path": path})
	_, _, err := NewRead().RunImages(context.Background(), string(args))
	if err == nil || strings.Contains(err.Error(), "binary") {
		t.Fatalf("err = %v", err)
	}
}
