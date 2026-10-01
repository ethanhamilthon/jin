package tools

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
)

const maxReadOutput = 32 << 10

const readSchema = `{"type":"function","function":{"name":"read","description":"Read a text file from disk, optionally a line range","parameters":{"type":"object","properties":{"path":{"type":"string","description":"File path to read"},"offset":{"type":"integer","description":"1-based line number to start from"},"limit":{"type":"integer","description":"Maximum number of lines to return"}},"required":["path"],"additionalProperties":false}}}`

type Read struct{}

func NewRead() Read { return Read{} }

func (Read) Name() string { return "read" }

func (Read) Schema() json.RawMessage { return json.RawMessage(readSchema) }

func (Read) Summary(argumentsJSON string) (string, bool) {
	args, ok := parseReadArgs(argumentsJSON)
	if !ok {
		return "", false
	}
	return args.summary(), true
}

func (Read) Run(_ context.Context, argumentsJSON string) (string, error) {
	args, ok := parseReadArgs(argumentsJSON)
	if !ok {
		return "", errors.New("invalid read tool arguments")
	}
	data, err := os.ReadFile(args.Path)
	if err != nil {
		return "", err
	}
	lines := strings.Split(string(data), "\n")
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
	return out.String(), nil
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
