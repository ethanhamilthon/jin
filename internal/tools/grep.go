package tools

import (
	"context"
	"encoding/json"
	"os"
)

const grepSchema = `{"type":"function","function":{"name":"grep","description":"Search file contents with a regular expression (Go syntax: no lookahead or backreferences). Searches the working directory unless path is given. Inside a git repository files that git ignores are skipped; elsewhere hidden folders, node_modules, vendor and dist are. Lines come back as ` + "`path:line:text`" + `; context lines use - instead of :. Output is cut at 32 KB: narrow it with path, glob or a more exact pattern.","parameters":{"type":"object","properties":{"pattern":{"type":"string","description":"Regular expression to look for"},"path":{"type":"string","description":"Folder or file to search; default is the working directory"},"glob":{"type":"string","description":"Only files matching this, such as *.go (file name) or cmd/*.go (path under the searched folder)"},"ignore_case":{"type":"boolean","description":"Match without regard to case"},"context":{"type":"integer","minimum":0,"maximum":10,"description":"Lines of context around each match; default 0"},"mode":{"type":"string","enum":["lines","files","count"],"description":"lines: matching lines (default); files: only paths with a match; count: matches per file"}},"required":["pattern"],"additionalProperties":false}}}`

type Grep struct{ dir string }

func (Grep) Name() string { return "grep" }

func (Grep) Schema() json.RawMessage { return json.RawMessage(grepSchema) }

func (Grep) Summary(argumentsJSON string) (string, bool) {
	args, err := parseGrepArgs(argumentsJSON)
	if err != nil {
		return "", false
	}
	return args.summary(), true
}

func (g Grep) Run(ctx context.Context, argumentsJSON string) (string, error) {
	args, err := parseGrepArgs(argumentsJSON)
	if err != nil {
		return "", err
	}
	re, err := args.regexp()
	if err != nil {
		return "", err
	}
	root := toolPath(g.dir, orDot(args.Path))
	info, err := os.Stat(root)
	if err != nil {
		return "", err
	}
	files, err := grepFiles(ctx, root, info, args)
	if err != nil {
		return "", err
	}
	result := &grepResult{mode: args.Mode, context: args.Context, skipped: map[string]int{}}
	for _, f := range files {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		result.add(f.shown, searchFile(re, f.abs))
	}
	return result.String(), nil
}

func orDot(path string) string {
	if path == "" {
		return "."
	}
	return path
}
