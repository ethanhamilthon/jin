package tools

import (
	"encoding/json"
	"errors"
	"path"
	"regexp"
	"strings"
)

type grepArgs struct {
	Pattern    string `json:"pattern"`
	Path       string `json:"path"`
	Glob       string `json:"glob"`
	Mode       string `json:"mode"`
	IgnoreCase bool   `json:"ignore_case"`
	Context    int    `json:"context"`
}

func parseGrepArgs(argumentsJSON string) (grepArgs, error) {
	var args grepArgs
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil || args.Pattern == "" {
		return args, errors.New("invalid grep tool arguments: pattern is required")
	}
	switch args.Mode {
	case "":
		args.Mode = "lines"
	case "lines", "files", "count":
	default:
		return args, errors.New("mode must be lines, files or count")
	}
	if args.Context < 0 || args.Context > 10 {
		return args, errors.New("context must be from 0 to 10")
	}
	if _, err := path.Match(args.Glob, ""); err != nil {
		return args, errors.New("invalid glob: " + args.Glob)
	}
	return args, nil
}

func (a grepArgs) regexp() (*regexp.Regexp, error) {
	expr := a.Pattern
	if a.IgnoreCase {
		expr = "(?i)" + expr
	}
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, errors.New("invalid pattern: " + err.Error())
	}
	return re, nil
}

func (a grepArgs) summary() string {
	if a.Path == "" {
		return a.Pattern
	}
	return a.Pattern + " in " + strings.TrimSuffix(a.Path, "/")
}

// globMatch compares a glob with a slash-separated path under the search
// root; a glob without a slash is compared with the file name only.
func globMatch(glob, rel string) bool {
	if glob == "" {
		return true
	}
	if !strings.Contains(glob, "/") {
		rel = path.Base(rel)
	}
	ok, _ := path.Match(glob, rel)
	return ok
}
