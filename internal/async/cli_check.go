package async

import (
	"errors"
	"fmt"
	"io"
	"jin/internal/store"
	"os"
	"strconv"
)

// check prints the state of a task, then its output: the last --limit
// characters, or everything without a limit.
func check(db *store.DB, opts options, stdout, stderr io.Writer) int {
	task, found, err := db.AsyncTask(opts.id)
	if err != nil {
		fmt.Fprintln(stderr, "jin async:", err)
		return 1
	}
	if !found {
		fmt.Fprintf(stderr, "jin async: no task %q\n", opts.id)
		return 1
	}
	status := "status: " + task.Status
	if task.Status != store.AsyncRunning {
		status += " exit: " + strconv.Itoa(task.ExitCode)
	}
	text, truncated, err := Tail(task.LogPath, opts.limit)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		fmt.Fprintln(stderr, "jin async:", err)
		return 1
	}
	fmt.Fprintln(stdout, status)
	if truncated {
		fmt.Fprintf(stdout, "[showing the last %d characters]\n", opts.limit)
	}
	if text == "" {
		fmt.Fprintln(stdout, "(no output yet)")
		return 0
	}
	fmt.Fprint(stdout, text)
	if text[len(text)-1] != '\n' {
		fmt.Fprintln(stdout)
	}
	return 0
}

type options struct {
	id, session string
	limit       int
	noNewline   bool
}

func parseArgs(args []string) (options, []string, error) {
	var opts options
	var words []string
	for i := 0; i < len(args); i++ {
		arg := args[i]
		value := func() (string, error) {
			if i+1 >= len(args) {
				return "", fmt.Errorf("%s needs a value", arg)
			}
			i++
			return args[i], nil
		}
		var err error
		switch arg {
		case "--id":
			opts.id, err = value()
		case "--session":
			opts.session, err = value()
		case "--limit":
			var v string
			if v, err = value(); err == nil {
				if opts.limit, err = strconv.Atoi(v); err != nil || opts.limit < 0 {
					err = fmt.Errorf("invalid --limit %q", v)
				}
			}
		case "--no-newline":
			opts.noNewline = true
		case "--":
			words = append(words, args[i+1:]...)
			return opts, words, nil
		default:
			words = append(words, arg)
		}
		if err != nil {
			return opts, nil, err
		}
	}
	return opts, words, nil
}
