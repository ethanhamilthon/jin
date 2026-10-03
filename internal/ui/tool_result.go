package ui

import (
	"encoding/json"
	"strconv"
	"strings"

	"jin/internal/diff"
	"jin/internal/provider"
	"jin/internal/tools"
)

// resultLines is how many lines of tool output the chat shows.
const resultLines = 5

// resultText is the body of a tool result entry: one line per row, each
// starting with its diff mark (' ' for plain output). The first line may be
// a note of how many lines were left out.
func resultText(tool, output string, changes []tools.Change) string {
	if tool == "bash" {
		return tailText(bashLines(output))
	}
	var lines []diff.Line
	for _, change := range changes {
		lines = append(lines, diff.Lines(change.Before, change.After)...)
	}
	if len(lines) == 0 {
		return tailText([]diff.Line{{Op: diff.Keep, Text: firstLine(output)}})
	}
	return tailText(lines)
}

func bashLines(output string) []diff.Line {
	var lines []diff.Line
	for _, line := range strings.Split(strings.TrimRight(output, "\n "), "\n") {
		lines = append(lines, diff.Line{Op: diff.Keep, Text: strings.TrimRight(line, "\r ")})
	}
	return lines
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
// answer. edit has its diff in the arguments; write only knows the new
// content, so it shows as all added.
func savedResultText(call provider.ToolCall, output string) (string, bool) {
	var args struct {
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
		Content   string `json:"content"`
	}
	switch call.Function.Name {
	case "bash":
		return resultText("bash", output, nil), true
	case "edit", "write":
		if strings.HasPrefix(output, "Error:") || json.Unmarshal([]byte(call.Function.Arguments), &args) != nil {
			return resultText(call.Function.Name, output, nil), true
		}
		change := tools.Change{Before: args.OldString, After: args.NewString}
		if call.Function.Name == "write" {
			change = tools.Change{After: args.Content}
		}
		return resultText(call.Function.Name, output, []tools.Change{change}), true
	}
	return "", false
}
