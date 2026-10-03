package headless

import (
	"errors"
	"flag"
	"fmt"
	"strconv"
	"strings"
	"time"
)

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
