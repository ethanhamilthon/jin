package cli

import (
	"fmt"
	"strings"
)

type sessionsOptions struct {
	all    bool
	format string
	words  []string
}

func parseSessionsArgs(args []string) (sessionsOptions, error) {
	opts := sessionsOptions{format: "text"}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--all":
			opts.all = true
		case arg == "--format":
			if i+1 >= len(args) {
				return opts, fmt.Errorf("missing argument for --format")
			}
			i++
			opts.format = args[i]
		case strings.HasPrefix(arg, "--format="):
			opts.format = strings.TrimPrefix(arg, "--format=")
		case strings.HasPrefix(arg, "-"):
			return opts, fmt.Errorf("unknown flag: %s", arg)
		default:
			opts.words = append(opts.words, arg)
		}
	}
	if opts.format != "text" && opts.format != "json" {
		return opts, fmt.Errorf("unknown format %q (use text or json)", opts.format)
	}
	return opts, nil
}
