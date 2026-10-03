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

const readSchema = `{"type":"function","function":{"name":"read","description":"Read a file from disk, optionally a line range. Pictures (png, jpeg, gif, webp, bmp) are attached for you to see","parameters":{"type":"object","properties":{"path":{"type":"string","description":"File path to read"},"offset":{"type":"integer","description":"1-based line number to start from"},"limit":{"type":"integer","description":"Maximum number of lines to return"}},"required":["path"],"additionalProperties":false}}}`

type Read struct{ seen *Seen }

func NewRead() Read { return Read{} }

// NewReadSeen shares seen with the other file tools of a session.
func NewReadSeen(seen *Seen) Read { return Read{seen: seen} }

func (Read) Name() string { return "read" }

func (Read) Schema() json.RawMessage { return json.RawMessage(readSchema) }

func (Read) Summary(argumentsJSON string) (string, bool) {
	args, ok := parseReadArgs(argumentsJSON)
	if !ok {
		return "", false
	}
	return args.summary(), true
}

func (r Read) Run(ctx context.Context, argumentsJSON string) (string, error) {
	text, _, err := r.RunImages(ctx, argumentsJSON)
	return text, err
}

// RunImages reads a picture as an attachment and anything else as text.
func (r Read) RunImages(_ context.Context, argumentsJSON string) (string, []Image, error) {
	args, ok := parseReadArgs(argumentsJSON)
	if !ok {
		return "", nil, errors.New("invalid read tool arguments")
	}
	data, err := os.ReadFile(args.Path)
	if err != nil {
		return "", nil, err
	}
	r.seen.Remember(args.Path)
	if picture, ok := loadImage(data); ok {
		return "Read image file [" + picture.MimeType + "]", []Image{picture}, nil
	}
	return readLines(string(data), args), nil, nil
}

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
