// Package headless runs jin without the TUI: one request in, one result out,
// for scripts and CI. It also holds the model-listing commands.
package headless

import (
	"errors"
	"flag"
	"fmt"
	"io"
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
	Cwd       string
	Provider  string
	MaxCost   float64
	MaxTurns  int
}

// Handles reports whether the arguments belong to a headless command, so the
// TUI is not started.
func Handles(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "models", "refresh-models", "projects":
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
	fs.StringVar(&opt.Cwd, "cwd", "", "")
	fs.StringVar(&opt.Provider, "provider", "", "")
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
	fs.Func("max-cost", "", func(v string) (err error) {
		opt.MaxCost, err = parseMaxCost(v)
		return err
	})
	fs.Func("max-turns", "", func(v string) (err error) {
		opt.MaxTurns, err = parseMaxTurns(v)
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
