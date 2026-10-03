package ui

import (
	"bytes"
	"encoding/hex"
	"jin/internal/paths"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// clipboardImagePNG extracts image data from the macOS clipboard via
// osascript, which triggers Cocoa's automatic pasteboard conversion to PNG
// regardless of the image's original representation (TIFF, JPEG, ...).
func clipboardImagePNG() ([]byte, bool) {
	info, err := exec.Command("osascript", "-e", "clipboard info").Output()
	if err != nil || !hasImageClass(string(info)) {
		return nil, false
	}
	out, err := exec.Command("osascript", "-e", `the clipboard as «class PNGf»`).Output()
	if err != nil {
		return nil, false
	}
	text := strings.TrimSpace(string(out))
	text = strings.TrimPrefix(text, "«data PNGf")
	text = strings.TrimSuffix(text, "»")
	data, err := hex.DecodeString(text)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return data, true
}

func hasImageClass(info string) bool {
	for _, class := range []string{"PNGf", "TIFF", "JPEG", "GIFf", "picture"} {
		if strings.Contains(info, class) {
			return true
		}
	}
	return false
}

func savePastedImage(data []byte) (string, error) {
	dir, err := paths.Global(".pasted")
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "paste-"+time.Now().Format("20060102-150405.000")+imageExt(data))
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// imageExt names a pasted picture by its bytes; Linux clipboards may hold
// JPEG or other formats where macOS always converts to PNG.
func imageExt(data []byte) string {
	switch {
	case bytes.HasPrefix(data, []byte("\xff\xd8\xff")):
		return ".jpg"
	case bytes.HasPrefix(data, []byte("GIF8")):
		return ".gif"
	case len(data) > 12 && string(data[:4]) == "RIFF" && string(data[8:12]) == "WEBP":
		return ".webp"
	case bytes.HasPrefix(data, []byte("BM")):
		return ".bmp"
	}
	return ".png"
}
