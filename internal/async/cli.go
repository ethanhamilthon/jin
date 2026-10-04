package async

import (
	"errors"
	"fmt"
	"io"
	"jin/internal/store"
)

// Main is `jin async ...`.
func Main(args []string, db *store.DB, version string, stdout, stderr io.Writer) int {
	fail := func(err error) int {
		fmt.Fprintln(stderr, "jin async:", err)
		return 1
	}
	if len(args) == 0 {
		fmt.Fprintln(stderr, "jin async: missing command (run, check, input, stop); see 'jin --help'")
		return 2
	}
	opts, words, err := parseArgs(args[1:])
	if err != nil {
		return fail(err)
	}
	switch args[0] {
	case "run":
		if len(words) != 1 {
			return fail(errors.New(`usage: jin async run "<command>" --session <session-id> [--stdin]`))
		}
		if opts.session == "" {
			return fail(errors.New("--session <session-id> is required (the Session id from your system prompt)"))
		}
		id, err := Start(version, opts.session, words[0], opts.stdin)
		if err != nil {
			return fail(err)
		}
		fmt.Fprintln(stdout, id)
	case "check":
		if opts.id == "" || len(words) > 0 {
			return fail(errors.New("usage: jin async check --id <task-id> [--limit <chars>]"))
		}
		return check(db, opts, stdout, stderr)
	case "input":
		if opts.id == "" || len(words) != 1 {
			return fail(errors.New(`usage: jin async input --id <task-id> "<text>" [--no-newline]`))
		}
		if err := Input(opts.id, words[0], opts.noNewline); err != nil {
			return fail(err)
		}
	case "stop":
		if opts.id == "" || len(words) > 0 {
			return fail(errors.New("usage: jin async stop --id <task-id>"))
		}
		if err := Stop(opts.id, StoppedByAgent); err != nil {
			return fail(err)
		}
		fmt.Fprintln(stdout, "stopped", opts.id)
	default:
		return fail(fmt.Errorf("unknown command %q (run, check, input, stop)", args[0]))
	}
	return 0
}
