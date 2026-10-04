//go:build unix

package async

import (
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"

	"jin/internal/tasklog"
)

func TestAdoptDirUsesExplicitSessionDirectory(t *testing.T) {
	db := startTestDaemon(t)
	storedDir, explicitDir := t.TempDir(), t.TempDir()
	if err := db.Touch("adopt-dir", storedDir, "model", "", "title"); err != nil {
		t.Fatal(err)
	}
	files, err := tasklog.New()
	if err != nil {
		t.Fatal(err)
	}
	log, err := os.OpenFile(files.Log, os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command("bash", "-c", "sleep 30")
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Stdout, cmd.Stderr = log, log
	if err := cmd.Start(); err != nil {
		log.Close()
		t.Fatal(err)
	}
	log.Close()
	t.Cleanup(func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) })
	go cmd.Wait()
	id, err := AdoptDir("test", "adopt-dir", explicitDir, "sleep 30", cmd.Process.Pid, cmd.Process.Pid, files.Log, files.Exit)
	if err != nil {
		t.Fatal(err)
	}
	task, found, err := db.AsyncTask(id)
	if err != nil || !found || task.Path != explicitDir || filepath.Clean(task.Path) != explicitDir {
		t.Fatalf("task path = %q, found=%v, err=%v; want %q", task.Path, found, err, explicitDir)
	}
	if err := Stop(id, StoppedByAgent); err != nil {
		t.Fatal(err)
	}
	waitEvent(t, db, explicitDir)
}
