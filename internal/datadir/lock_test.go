//go:build unix

package datadir

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"testing"
)

// otherProcess takes a lock the way a second jin process would: flock locks
// belong to an open file, so a second open conflicts even in this process.
func otherProcess(t *testing.T, dir string, how int) func() {
	t.Helper()
	file, err := os.OpenFile(filepath.Join(dir, lockName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(file.Fd()), how|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	return func() { file.Close() }
}

func TestMoveIsRefusedWhileOthersHoldTheFolder(t *testing.T) {
	dir := t.TempDir()
	if err := holdAt(dir); err != nil {
		t.Fatal(err)
	}
	defer Release()
	release := otherProcess(t, dir, syscall.LOCK_SH)
	if err := CheckAlone(); !errors.Is(err, ErrInUse) {
		t.Fatalf("check with another process = %v", err)
	}
	if err := Exclusive(); !errors.Is(err, ErrInUse) {
		t.Fatalf("exclusive with another process = %v", err)
	}
	release()
	if err := CheckAlone(); err != nil {
		t.Fatalf("check alone = %v", err)
	}
	release = otherProcess(t, dir, syscall.LOCK_SH)
	release()
	if err := Exclusive(); err != nil {
		t.Fatalf("exclusive alone = %v", err)
	}
}

func TestHoldFailsDuringAMove(t *testing.T) {
	dir := t.TempDir()
	release := otherProcess(t, dir, syscall.LOCK_EX)
	defer release()
	if err := holdAt(dir); err == nil {
		Release()
		t.Fatal("hold must fail while the folder is being moved")
	}
}

func TestSwapRollsBackAFailedLastStep(t *testing.T) {
	root := t.TempDir()
	current, other := filepath.Join(root, ".jin"), filepath.Join(root, "other")
	seed(t, current, "current")
	seed(t, other, "other")
	calls := 0
	rename = func(from, to string) error {
		if calls++; calls == 3 {
			return errors.New("disk full")
		}
		return os.Rename(from, to)
	}
	defer func() { rename = os.Rename }()
	if err := Swap(current, other); err == nil {
		t.Fatal("the failed rename must be reported")
	}
	if marker(current) != "current" || marker(other) != "other" {
		t.Fatalf("not rolled back: %q %q", marker(current), marker(other))
	}
}
