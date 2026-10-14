package session

import (
	"jin/internal/core"
	"jin/internal/provider"
)

// pictureEntries shows the pictures of a tool-result message. paths are the
// files the read calls of the batch opened, in order; with fewer paths than
// pictures (older sessions) the rest show without a file name.
func pictureEntries(msg provider.Message, paths []string) []Entry {
	entries := make([]Entry, len(msg.Images))
	for i := range msg.Images {
		path := ""
		if i < len(paths) {
			path = paths[i]
		}
		entries[i] = pictureEntry(core.PictureLabel(path, msg.Images[i]), &msg.Images[i])
	}
	return entries
}
