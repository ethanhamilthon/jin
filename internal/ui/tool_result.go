package ui

import (
	"strconv"
	"strings"

	"jin/internal/diff"
	"jin/internal/provider"
	"jin/internal/session"
	"jin/internal/tools"
)

// resultLines is how many lines of tool output the chat shows.
const resultLines = 5

// resultText is the body of a tool result entry: one line per row, each
// starting with its diff mark (' ' for plain output). The first line may be
// a note of how many lines were left out.
func resultText(tool, output string, changes []tools.Change) string {
	return tailText(session.ResultLines(tool, output, changes))
}

const hiddenMark = '…'

func tailText(lines []diff.Line) string {
	var b strings.Builder
	if hidden := len(lines) - resultLines; hidden > 0 {
		b.WriteRune(hiddenMark)
		b.WriteString(strconv.Itoa(hidden) + " more lines\n")
		lines = lines[hidden:]
	}
	for _, line := range lines {
		b.WriteByte(byte(line.Op))
		b.WriteString(line.Text)
		b.WriteByte('\n')
	}
	return strings.TrimSuffix(b.String(), "\n")
}

// savedResultText rebuilds a result entry from a stored tool call and its
// answer.
func savedResultText(call provider.ToolCall, output string) (string, bool) {
	lines, ok := session.SavedResultLines(call, output)
	if !ok {
		return "", false
	}
	return tailText(lines), true
}
