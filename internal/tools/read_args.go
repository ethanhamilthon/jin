package tools

import (
	"encoding/json"
	"strconv"
	"strings"
)

func readLines(content string, args readArgs) string {
	lines := strings.Split(content, "\n")
	start := 0
	if args.Offset > 1 {
		start = min(args.Offset-1, len(lines))
	}
	end := len(lines)
	if args.Limit > 0 {
		end = min(start+args.Limit, len(lines))
	}
	var out strings.Builder
	for i := start; i < end; i++ {
		out.WriteString(strconv.Itoa(i + 1))
		out.WriteByte('\t')
		out.WriteString(lines[i])
		out.WriteByte('\n')
		if out.Len() > maxReadOutput {
			out.WriteString("[output truncated]")
			break
		}
	}
	return out.String()
}

type readArgs struct {
	Path   string `json:"path"`
	Offset int    `json:"offset"`
	Limit  int    `json:"limit"`
}

func (a readArgs) summary() string {
	if a.Offset <= 0 && a.Limit <= 0 {
		return a.Path
	}
	if a.Limit <= 0 {
		return a.Path + ":" + strconv.Itoa(a.Offset)
	}
	offset := max(a.Offset, 1)
	return a.Path + ":" + strconv.Itoa(offset) + "-" + strconv.Itoa(offset+a.Limit-1)
}

func parseReadArgs(argumentsJSON string) (readArgs, bool) {
	var args readArgs
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil || strings.TrimSpace(args.Path) == "" {
		return readArgs{}, false
	}
	return args, true
}
