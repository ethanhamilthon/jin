// Package headless runs jin without the TUI: one request in, one result out,
// for scripts and CI. It also holds the model-listing commands.
package headless

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Options is what `jin -p` was asked to do.
type Options struct {
	Prompt    string
	Format    string
	Continue  bool
	Session   string
	NoSession bool
	Model     string
	Effort    string
	Tools     []string
	ToolsSet  bool
	Exclude   []string
	NoTools   bool
	Timeout   time.Duration
}

// Handles reports whether the arguments belong to a headless command, so the
// TUI is not started.
func Handles(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "models", "refresh-models":
		return true
	}
	for _, arg := range args {
		if arg == "--" {
			return false
		}
		if arg == "-p" || arg == "--p" {
			return true
		}
	}
	return false
}

// parseInterleaved parses flags that may come before, between and after the
// positional words; everything after "--" is positional.
func parseInterleaved(fs *flag.FlagSet, args []string) ([]string, error) {
	var tail []string
	for i, arg := range args {
		if arg == "--" {
			args, tail = args[:i], args[i+1:]
			break
		}
	}
	var positional []string
	for len(args) > 0 {
		if err := fs.Parse(args); err != nil {
			return nil, err
		}
		args = fs.Args()
		if len(args) == 0 {
			break
		}
		positional = append(positional, args[0])
		args = args[1:]
	}
	return append(positional, tail...), nil
}

func splitNames(value string) ([]string, error) {
	var names []string
	for _, name := range strings.Split(value, ",") {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, errors.New("empty tool name in list")
		}
		names = append(names, name)
	}
	return names, nil
}

// parseTimeout accepts a Go duration ("90s", "10m") or a bare number of
// seconds.
func parseTimeout(value string) (time.Duration, error) {
	if n, err := strconv.ParseFloat(value, 64); err == nil {
		value = strconv.FormatFloat(n, 'f', -1, 64) + "s"
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return 0, fmt.Errorf("invalid timeout %q", value)
	}
	return d, nil
}

func ParseArgs(args []string) (Options, error) {
	var opt Options
	fs := flag.NewFlagSet("jin", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.Bool("p", false, "")
	fs.BoolVar(&opt.Continue, "continue", false, "")
	fs.BoolVar(&opt.Continue, "c", false, "")
	fs.StringVar(&opt.Session, "session", "", "")
	fs.BoolVar(&opt.NoSession, "no-session", false, "")
	fs.StringVar(&opt.Model, "model", "", "")
	fs.StringVar(&opt.Effort, "effort", "", "")
	fs.StringVar(&opt.Format, "format", "text", "")
	fs.BoolVar(&opt.NoTools, "no-tools", false, "")
	fs.Func("tools", "", func(v string) (err error) {
		opt.ToolsSet = true
		opt.Tools, err = splitNames(v)
		return err
	})
	fs.Func("exclude-tools", "", func(v string) (err error) {
		opt.Exclude, err = splitNames(v)
		return err
	})
	fs.Func("timeout", "", func(v string) (err error) {
		opt.Timeout, err = parseTimeout(v)
		return err
	})
	words, err := parseInterleaved(fs, args)
	if err != nil {
		return Options{}, err
	}
	opt.Prompt = strings.Join(words, " ")
	if opt.Format != "text" && opt.Format != "json" {
		return Options{}, fmt.Errorf("unknown format %q (use text or json)", opt.Format)
	}
	if opt.Continue && opt.Session != "" {
		return Options{}, errors.New("--continue and --session cannot be used together")
	}
	return opt, nil
}
