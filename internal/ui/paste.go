package ui

import (
	"github.com/clipperhouse/displaywidth"
)

// pasteClipboard reads the system clipboard: an image is saved under
// the global data directory (.pasted) and its path is returned as text; otherwise the clipboard's
// plain text is returned as-is. ok is false only when the clipboard could
// not be read at all.
func pasteClipboard() (text string, ok bool) {
	text, _, ok = readClipboard()
	return text, ok
}

// readClipboard is pasteClipboard that also tells whether text is the path
// of a saved image.
func readClipboard() (text string, image, ok bool) {
	if data, found := clipboardImage(); found {
		if path, err := savePastedImage(data); err == nil {
			return path, true, true
		}
	}
	text, ok = clipboardText()
	return text, false, ok
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
