//go:build unix

package async

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"jin/internal/store"
)

const (
	idleExit   = 10 * time.Minute
	killGrace  = 3 * time.Second
	connLimit  = 30 * time.Second
	requestMax = 8 << 20
)

type task struct {
	stdin io.WriteCloser
}

type daemon struct {
	db      *store.DB
	version string
	mu      sync.Mutex
	tasks   map[string]*task
	active  time.Time
	done    chan struct{}
}

// Serve runs the daemon until it has been idle for ten minutes. A second
// daemon finds the lock taken and returns at once.
func Serve(ctx context.Context, db *store.DB, version string) error {
	dir, err := dataDir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	lock, err := lockPath()
	if err != nil {
		return err
	}
	lockFile, err := os.OpenFile(lock, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	defer lockFile.Close()
	if err := syscall.Flock(int(lockFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return nil
		}
		return err
	}
	sock, err := socketPath()
	if err != nil {
		return err
	}
	_ = os.Remove(sock)
	listener, err := net.Listen("unix", sock)
	if err != nil {
		return err
	}
	defer os.Remove(sock)
	if err := os.Chmod(sock, 0o600); err != nil {
		listener.Close()
		return err
	}
	d := &daemon{db: db, version: version, tasks: map[string]*task{}, active: time.Now(), done: make(chan struct{})}
	d.recover()
	cleanOldFiles(dir)
	go d.accept(listener)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			listener.Close()
			return nil
		case <-d.done:
			listener.Close()
			return nil
		case <-ticker.C:
			if d.idleFor() > idleExit {
				listener.Close()
				return nil
			}
		}
	}
}

func (d *daemon) idleFor() time.Duration {
	d.mu.Lock()
	defer d.mu.Unlock()
	if len(d.tasks) > 0 {
		return 0
	}
	return time.Since(d.active)
}

func (d *daemon) touch() {
	d.mu.Lock()
	d.active = time.Now()
	d.mu.Unlock()
}

func (d *daemon) accept(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go d.serve(conn)
	}
}

func (d *daemon) serve(conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(connLimit))
	var req request
	if err := json.NewDecoder(io.LimitReader(conn, requestMax)).Decode(&req); err != nil {
		return
	}
	d.touch()
	resp := d.handle(req)
	resp.Version = d.version
	_ = json.NewEncoder(conn).Encode(resp)
}

func (d *daemon) handle(req request) response {
	switch req.Op {
	case opHello:
		d.mu.Lock()
		idle := len(d.tasks) == 0
		d.mu.Unlock()
		return response{OK: true, Idle: idle}
	case opShutdown:
		d.mu.Lock()
		idle := len(d.tasks) == 0
		d.mu.Unlock()
		if !idle {
			return response{Error: "daemon has running tasks"}
		}
		go func() { time.Sleep(100 * time.Millisecond); close(d.done) }()
		return response{OK: true}
	case opRun:
		id, err := d.run(req)
		if err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true, ID: id}
	case opAdopt:
		id, err := d.adopt(req)
		if err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true, ID: id}
	case opInput:
		if err := d.input(req); err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true}
	case opStop:
		if err := d.stop(req); err != nil {
			return response{Error: err.Error()}
		}
		return response{OK: true}
	}
	return response{Error: "unknown operation " + strconv.Quote(req.Op)}
}

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
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
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

// wait ends a task when its process does. A task that was stopped has
// already been closed and announced by stop, so nothing more is reported.
func (d *daemon) wait(id string, cmd *exec.Cmd, session, path, logPath string) {
	err := cmd.Wait()
	code := exitCode(cmd, err)
	d.mu.Lock()
	delete(d.tasks, id)
	d.active = time.Now()
	d.mu.Unlock()
	status := store.AsyncDone
	if code != 0 {
		status = store.AsyncFailed
	}
	won, dbErr := d.db.FinishAsyncTask(id, status, code)
	if dbErr != nil || !won {
		return
	}
	output, truncated, _ := Tail(logPath, eventTail)
	if truncated {
		output = fmt.Sprintf("[output cut: last %d characters; run `jin async check --id %s` for more]\n%s", eventTail, id, output)
	}
	_ = d.db.AddAsyncEvent(session, path, ResultText(id, status, code, output))
}

func exitCode(cmd *exec.Cmd, err error) int {
	if cmd.ProcessState == nil {
		if err != nil {
			return 1
		}
		return 0
	}
	if code := cmd.ProcessState.ExitCode(); code >= 0 {
		return code
	}
	if ws, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && ws.Signaled() {
		return 128 + int(ws.Signal())
	}
	return 1
}

func (d *daemon) input(req request) error {
	d.mu.Lock()
	t := d.tasks[req.ID]
	d.mu.Unlock()
	if t != nil && t.stdin == nil {
		return fmt.Errorf("task %s was started by the bash tool and has no stdin you can write to", req.ID)
	}
	if t == nil {
		return fmt.Errorf("task %s is not running", req.ID)
	}
	text := req.Text
	if !req.NoNewline {
		text += "\n"
	}
	_, err := io.WriteString(t.stdin, text)
	return err
}

// stop closes the task first, so the wait goroutine stays quiet, then
// announces it and ends the process group: SIGTERM, and SIGKILL after a
// grace period.
func (d *daemon) stop(req request) error {
	rec, found, err := d.db.AsyncTask(req.ID)
	if err != nil {
		return err
	}
	if !found || rec.Status != store.AsyncRunning {
		return fmt.Errorf("task %s is not running", req.ID)
	}
	won, err := d.db.FinishAsyncTask(rec.ID, store.AsyncStopped, 143)
	if err != nil {
		return err
	}
	if !won {
		return fmt.Errorf("task %s is not running", req.ID)
	}
	note := "stopped by the agent"
	if req.By == StoppedByUser {
		note = "stopped manually by the user"
	}
	_ = d.db.AddAsyncEvent(rec.SessionID, rec.Path, ResultText(rec.ID, store.AsyncStopped, -1, note))
	killGroup(rec.PGID)
	return nil
}

// killGroup sends SIGTERM to a process group and SIGKILL when it is still
// there after the grace period.
func killGroup(pgid int) {
	if pgid <= 1 {
		return
	}
	_ = syscall.Kill(-pgid, syscall.SIGTERM)
	go func() {
		time.Sleep(killGrace)
		if syscall.Kill(-pgid, 0) == nil {
			_ = syscall.Kill(-pgid, syscall.SIGKILL)
		}
	}()
}

// recover settles tasks a previous daemon left behind: their supervisor is
// gone, so nobody would announce them. A dead process is marked failed; a
// live group that still belongs to the task is ended.
func (d *daemon) recover() {
	tasks, err := d.db.RunningAsyncTasks("")
	if err != nil {
		return
	}
	for _, t := range tasks {
		if t.ExitPath != "" {
			d.watch(t)
			continue
		}
		alive := t.PID > 0 && syscall.Kill(t.PID, 0) == nil && sameProcess(t)
		note := "the async daemon restarted and lost this task"
		if alive {
			killGroup(t.PGID)
			note += "; its process was stopped"
		}
		if won, err := d.db.FinishAsyncTask(t.ID, store.AsyncFailed, 1); err == nil && won {
			_ = d.db.AddAsyncEvent(t.SessionID, t.Path, ResultText(t.ID, store.AsyncFailed, 1, note))
		}
	}
}

// sameProcess guards against a reused pid: the start time recorded with the
// task must still match. An unknown stamp (0) counts as a match.
func sameProcess(t store.AsyncTask) bool {
	if t.ProcStartedAt == 0 {
		return true
	}
	now := procStamp(t.PID)
	return now == 0 || now == t.ProcStartedAt
}

// procStamp identifies a process by its start time. It is 0 when unknown.
func procStamp(pid int) int64 {
	out, err := exec.Command("ps", "-o", "lstart=", "-p", strconv.Itoa(pid)).Output()
	if err != nil || len(strings.TrimSpace(string(out))) == 0 {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(strings.TrimSpace(string(out))))
	return int64(h.Sum64() >> 1)
}

// startDetached starts `jin daemon` in its own session with no terminal, so
// it survives the shell, the TUI and the terminal that asked for it.
func startDetached(exe string) error {
	cmd := exec.Command(exe, "daemon")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}
