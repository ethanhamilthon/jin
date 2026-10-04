//go:build unix

package datadir

import (
	"errors"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

const lockName = ".jin.lock"

// exclusiveTries covers a daemon that was just told to stop and still exits.
const exclusiveTries = 10

// ErrInUse is the refusal to move a data directory that others still use.
var ErrInUse = errors.New("other jin processes use the data folder; close other jin windows and async tasks first")

var held *os.File

// Hold takes a shared lock on the data directory for the life of this
// process, so a move of the directory can tell that it is in use.
func Hold() error {
	dir, err := Current()
	if err != nil {
		return err
	}
	return holdAt(dir)
}

func holdAt(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	file, err := os.OpenFile(filepath.Join(dir, lockName), os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return err
	}
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH|syscall.LOCK_NB); err != nil {
		file.Close()
		if errors.Is(err, syscall.EWOULDBLOCK) {
			return errors.New("the jin data folder is being moved; try again in a moment")
		}
		return err
	}
	held = file
	return nil
}

// Release drops the lock of Hold.
func Release() {
	if held != nil {
		held.Close()
		held = nil
	}
}

// Exclusive turns the lock of this process into an exclusive one for a
// move. It does not wait for others: it fails with ErrInUse.
func Exclusive() error {
	if held == nil {
		return nil
	}
	for attempt := 1; ; attempt++ {
		err := syscall.Flock(int(held.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			return nil
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) {
			return err
		}
		if attempt == exclusiveTries {
			_ = syscall.Flock(int(held.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
			return ErrInUse
		}
		time.Sleep(100 * time.Millisecond)
	}
}

// CheckAlone fails with ErrInUse while other jin processes use the data
// directory.
func CheckAlone() error {
	if err := Exclusive(); err != nil {
		return err
	}
	if held == nil {
		return nil
	}
	return syscall.Flock(int(held.Fd()), syscall.LOCK_SH|syscall.LOCK_NB)
}
