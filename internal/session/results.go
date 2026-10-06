package session

import (
	"encoding/json"
	"strings"

	"jin/internal/diff"
	"jin/internal/provider"
	"jin/internal/tools"
)

// ResultLines is the output of a bash call, or the diff of the files an
// edit or write call changed; otherwise the first line of the output.
func ResultLines(tool, output string, changes []tools.Change) []diff.Line {
	if tool == "bash" {
		return bashLines(output)
	}
	var lines []diff.Line
	for _, change := range changes {
		lines = append(lines, diff.Lines(change.Before, change.After)...)
	}
	if len(lines) == 0 {
		return []diff.Line{{Op: diff.Keep, Text: FirstLine(output)}}
	}
	return lines
}

func bashLines(output string) []diff.Line {
	var lines []diff.Line
	for _, line := range strings.Split(strings.TrimRight(output, "\n "), "\n") {
		lines = append(lines, diff.Line{Op: diff.Keep, Text: strings.TrimRight(line, "\r ")})
	}
	return lines
}

// SavedResultLines rebuilds the lines of a stored bash, edit or write call
// and its answer. edit has its diff in the arguments; write only knows the
// new content, so it shows as all added.
func SavedResultLines(call provider.ToolCall, output string) ([]diff.Line, bool) {
	var args struct {
		OldString string `json:"old_string"`
		NewString string `json:"new_string"`
		Content   string `json:"content"`
	}
	switch call.Function.Name {
	case "bash":
		return ResultLines("bash", output, nil), true
	case "edit", "write":
		if strings.HasPrefix(output, "Error:") || json.Unmarshal([]byte(call.Function.Arguments), &args) != nil {
			return ResultLines(call.Function.Name, output, nil), true
		}
		change := tools.Change{Before: args.OldString, After: args.NewString}
		if call.Function.Name == "write" {
			change = tools.Change{After: args.Content}
		}
		return ResultLines(call.Function.Name, output, []tools.Change{change}), true
	}
	return nil, false
}
