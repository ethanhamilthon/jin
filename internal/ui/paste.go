package ui

import (
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/clipperhouse/displaywidth"

	"jin/internal/paths"
)

// pasteClipboard reads the system clipboard: an image is saved under
// the global data directory (.pasted) and its path is returned as text; otherwise the clipboard's
// plain text is returned as-is. ok is false only when the clipboard could
// not be read at all.
func pasteClipboard() (text string, ok bool) {
	if data, found := clipboardImagePNG(); found {
		if path, err := savePastedImage(data); err == nil {
			return path, true
		}
	}
	out, err := exec.Command("pbpaste").Output()
	if err != nil {
		return "", false
	}
	return string(out), true
}

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
	path := filepath.Join(dir, "paste-"+time.Now().Format("20060102-150405.000")+".png")
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return "", err
	}
	return path, nil
}

// insertClusters splices text into a chat-style cluster input at cursor,
// advancing cursor past the inserted content.
func insertClusters(input *[]string, cursor *int, text string) {
	graphemes := displaywidth.StringGraphemes(text)
	for graphemes.Next() {
		cluster := graphemes.Value()
		*input = append(*input, "")
		copy((*input)[*cursor+1:], (*input)[*cursor:])
		(*input)[*cursor] = cluster
		*cursor++
	}
}
