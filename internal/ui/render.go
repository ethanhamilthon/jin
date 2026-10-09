package ui

import (
	"github.com/gdamore/tcell/v3"

	"jin/internal/core"
	"jin/internal/provider"
)

type chatEntry struct {
	kind core.UpdateKind
	text string
	tool string
	// pending lists the #prompts of a promptsEntry whose commands still run.
	pending []string
	// image is the picture an info entry shows under its label.
	image *provider.Image
}

type chatSpan struct {
	text  string
	style tcell.Style
	// spin marks a span that shows the current spinner frame when drawn,
	// so the rows need no rebuilding while it turns.
	spin bool
	// shimmer draws the span cell by cell in the moving logo gradient.
	shimmer bool
	// lineBreak ends the row here: a line break inside a markdown paragraph.
	lineBreak bool
}

type chatRow struct {
	kind     core.UpdateKind
	text     string
	spans    []chatSpan
	fill     tcell.Style
	hasFill  bool
	fillWide bool
}

func (r chatRow) blank() bool { return r.text == "" && len(r.spans) == 0 }

func entryRows(entry chatEntry, width int) []chatRow {
	if entry.image != nil {
		return pictureRows(entry, width)
	}
	if entry.tool == sectionEntry {
		return sectionRows(entry.text, width)
	}
	if entry.tool == logoEntry {
		return logoRows(entry.text)
	}
	if entry.tool == taskEntry {
		return taskRows(entry.text, width)
	}
	if entry.tool == promptsEntry {
		return promptsRows(entry, width)
	}
	switch entry.kind {
	case core.UpdateAsk:
		return sectionRows(entry.text, width)
	case core.UpdateCompacted:
		return dividerRows(entry.text, width)
	case core.UpdateAssistant:
		return append(markdownRows(entry.text, width-2), chatRow{})
	case core.UpdateReasoning:
		return reasoningRows(entry.text, width)
	case core.UpdateToolCall:
		return toolCallRows(entry.tool, entry.text, width)
	case core.UpdateToolResult:
		return toolResultRows(entry.text, width)
	case core.UpdateUser:
		return append(userRows(entry.text, width), chatRow{})
	}
	var rows []chatRow
	for _, line := range wrapChat(entry.text, width-2) {
		rows = append(rows, chatRow{kind: entry.kind, text: line})
	}
	return append(rows, chatRow{})
}
