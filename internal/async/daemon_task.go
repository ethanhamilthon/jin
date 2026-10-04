//go:build unix

package async

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"jin/internal/store"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
)

func newTaskID() string {
	var buf [4]byte
	_, _ = rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func (d *daemon) run(req request) (string, error) {
	command := strings.TrimSpace(req.Command)
	if command == "" {
		return "", errors.New("empty command")
	}
	if req.Session == "" {
		return "", errors.New("session id is required")
	}
	path := req.Cwd
	if rec, ok, err := d.db.GetSession(req.Session); err == nil && ok {
		path = rec.Path
	}
	id := newTaskID()
	logPath, err := LogPath(id)
	if err != nil {
		return "", err
	}
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return "", err
	}
	defer logFile.Close()
	cmd := exec.Command("bash", "-c", command)
	if info, err := os.Stat(req.Cwd); err == nil && info.IsDir() {
		cmd.Dir = req.Cwd
	}
	// The task gets the environment of whoever asked for it, JIN_DEPTH
	// included, never the daemon's own.
	cmd.Env = req.Env
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Stdout, cmd.Stderr = logFile, logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	var stdin io.WriteCloser
	if req.Stdin {
		if stdin, err = cmd.StdinPipe(); err != nil {
			return "", err
		}
	}
	if err := cmd.Start(); err != nil {
		return "", err
	}
	pid := cmd.Process.Pid
	record := store.AsyncTask{
		ID: id, SessionID: req.Session, Path: path, Command: command,
		PID: pid, PGID: pid, ProcStartedAt: procStamp(pid), LogPath: logPath, StartedAt: time.Now(),
	}
	if err := d.db.AddAsyncTask(record); err != nil {
		_ = syscall.Kill(-pid, syscall.SIGKILL)
		_ = cmd.Wait()
		return "", err
	}
	d.mu.Lock()
	d.tasks[id] = &task{stdin: stdin}
	d.mu.Unlock()
	go d.wait(id, cmd, req.Session, path, logPath)
	return id, nil
}
