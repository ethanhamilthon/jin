// Package session runs conversations for any front end: it turns the
// updates of an agent and the saved history into entries to show.
package session

import (
	"jin/internal/core"
	"jin/internal/diff"
	"jin/internal/provider"
)

// CompactedLabel marks the place where a summary replaced older messages.
const CompactedLabel = "Compacted"

// TaskTool marks an info entry that shows a background task result.
const TaskTool = "task"

// Entry is one block of a conversation as a front end shows it. Tool
// results carry their output or diff as Lines.
type Entry struct {
	Kind  core.UpdateKind `json:"kind"`
	Tool  string          `json:"tool,omitempty"`
	Text  string          `json:"text"`
	Lines []Line          `json:"lines,omitempty"`
	// Picture marks an entry that shows Image; the bytes stay out of the JSON
	// and are served by Manager.EntryImage.
	Picture bool            `json:"picture,omitempty"`
	Image   *provider.Image `json:"-"`
	// Pictures counts the images a user message carries, kept in Images.
	Pictures int              `json:"pictures,omitempty"`
	Images   []provider.Image `json:"-"`
}

func userEntry(text string, images []provider.Image) Entry {
	return Entry{Kind: core.UpdateUser, Text: text, Pictures: len(images), Images: images}
}

func pictureEntry(label string, image *provider.Image) Entry {
	return Entry{Kind: core.UpdateInfo, Text: label, Picture: true, Image: image}
}

// Line is one line of tool output: Op is " ", "+" or "-".
type Line struct {
	Op   string `json:"op"`
	Text string `json:"text"`
}

func toLines(lines []diff.Line) []Line {
	out := make([]Line, len(lines))
	for i, line := range lines {
		out[i] = Line{Op: string(rune(line.Op)), Text: line.Text}
	}
	return out
}

// DiffLines turns Lines back into diff lines.
func DiffLines(lines []Line) []diff.Line {
	out := make([]diff.Line, len(lines))
	for i, line := range lines {
		op := diff.Keep
		if line.Op != "" {
			op = diff.Op(line.Op[0])
		}
		out[i] = diff.Line{Op: op, Text: line.Text}
	}
	return out
}
