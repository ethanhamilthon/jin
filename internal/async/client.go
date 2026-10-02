package async

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"time"

	"jin/internal/store"
)

const (
	dialTimeout  = 500 * time.Millisecond
	startTimeout = 2 * time.Second
	callTimeout  = 15 * time.Second
)

func call(req request) (response, error) {
	sock, err := socketPath()
	if err != nil {
		return response{}, err
	}
	conn, err := net.DialTimeout("unix", sock, dialTimeout)
	if err != nil {
		return response{}, err
	}
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(callTimeout))
	if err := json.NewEncoder(conn).Encode(req); err != nil {
		return response{}, err
	}
	var resp response
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return response{}, err
	}
	if !resp.OK {
		return resp, errors.New(resp.Error)
	}
	return resp, nil
}

// ensureDaemon makes sure a daemon of this version answers. A daemon of
// another version is replaced when it has no running task; otherwise it keeps
// serving until it is idle.
func ensureDaemon(version string) error {
	resp, err := call(request{Op: opHello})
	if err == nil && resp.Version == version {
		return nil
	}
	if err == nil {
		if !resp.Idle {
			return nil
		}
		if _, err := call(request{Op: opShutdown}); err != nil {
			return nil
		}
		waitGone()
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	if err := startDetached(exe); err != nil {
		return err
	}
	deadline := time.Now().Add(startTimeout)
	for time.Now().Before(deadline) {
		if _, err := call(request{Op: opHello}); err == nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}
	return errors.New("the async daemon did not start")
}

func waitGone() {
	deadline := time.Now().Add(startTimeout)
	for time.Now().Before(deadline) {
		if _, err := call(request{Op: opHello}); err != nil {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
}

// Start runs a command in the background for a session and returns the task id.
func Start(version, session, command string) (string, error) {
	cwd, _ := os.Getwd()
	if err := ensureDaemon(version); err != nil {
		return "", err
	}
	resp, err := call(request{Op: opRun, Session: session, Cwd: cwd, Command: command, Env: os.Environ()})
	return resp.ID, err
}

// Adopt hands a running process of the bash tool to the daemon and returns
// the task id. The process, its log and its exit file stay where they are.
func Adopt(version, session, command string, pid, pgid int, log, exit string) (string, error) {
	cwd, _ := os.Getwd()
	if err := ensureDaemon(version); err != nil {
		return "", err
	}
	resp, err := call(request{Op: opAdopt, Session: session, Cwd: cwd, Command: command, PID: pid, PGID: pgid, Log: log, Exit: exit})
	return resp.ID, err
}

// Stop ends a task. by is StoppedByUser or StoppedByAgent.
func Stop(id, by string) error {
	_, err := call(request{Op: opStop, ID: id, By: by})
	return err
}

// Input writes a line to the stdin of a running task.
func Input(id, text string, noNewline bool) error {
	_, err := call(request{Op: opInput, ID: id, Text: text, NoNewline: noNewline})
	return err
}

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
			return fail(errors.New(`usage: jin async run "<command>" --session <session-id>`))
		}
		if opts.session == "" {
			return fail(errors.New("--session <session-id> is required (the Session id from your system prompt)"))
		}
		id, err := Start(version, opts.session, words[0])
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
