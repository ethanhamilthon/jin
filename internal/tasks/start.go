package tasks

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"time"

	"jin/internal/tasklog"
)

// Start runs command in dir as a new task of owner. With stdin the task gets
// a pipe that Input writes to; otherwise it reads end of file at once.
func (m *Manager) Start(owner, command, dir string, stdin bool) (Info, error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return Info{}, errors.New("empty command")
	}
	files, err := tasklog.New()
	if err != nil {
		return Info{}, err
	}
	logFile, err := os.OpenFile(files.Log, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return Info{}, err
	}
	defer logFile.Close()
	cmd := exec.Command("bash", "-c", command)
	cmd.Dir = dir
	cmd.Stdout, cmd.Stderr = logFile, logFile
	setGroup(cmd)
	var pipe *os.File
	if stdin {
		reader, writer, err := os.Pipe()
		if err != nil {
			return Info{}, err
		}
		defer reader.Close()
		cmd.Stdin, pipe = reader, writer
	}
	if err := cmd.Start(); err != nil {
		if pipe != nil {
			_ = pipe.Close()
		}
		tasklog.Remove(files.Log)
		return Info{}, err
	}
	finished := make(chan error, 1)
	go func() { finished <- cmd.Wait() }()
	return m.add(files, owner, command, dir, cmd, finished, pipe), nil
}

// Adopt takes over a command the bash tool started and is waiting on:
// finished delivers its Wait result.
func (m *Manager) Adopt(owner, command, dir string, files tasklog.Files, cmd *exec.Cmd, finished <-chan error) Info {
	return m.add(files, owner, command, dir, cmd, finished, nil)
}

func (m *Manager) add(files tasklog.Files, owner, command, dir string, cmd *exec.Cmd, finished <-chan error, stdin *os.File) Info {
	t := &task{
		info:    Info{ID: files.ID, Owner: owner, Command: command, Dir: dir, Log: files.Log, Status: Running, Exit: -1, Started: time.Now(), Stdin: stdin != nil},
		cmd:     cmd,
		stdin:   stdin,
		writing: make(chan struct{}, 1),
		ended:   make(chan struct{}),
	}
	m.mu.Lock()
	m.tasks[t.info.ID] = t
	m.mu.Unlock()
	info := t.info
	go m.watch(t, finished)
	return info
}

func (m *Manager) watch(t *task, finished <-chan error) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case err := <-finished:
			m.finish(t, err)
			return
		case <-ticker.C:
			_ = tasklog.Trim(t.info.Log)
		}
	}
}
