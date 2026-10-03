package async

import (
	"encoding/json"
	"errors"
	"net"
	"os"
	"time"
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
