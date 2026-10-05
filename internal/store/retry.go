package store

import (
	"strings"
	"time"
)

// busyRetries is how often a write is tried while another process holds the lock.
const busyRetries = 5

// retryBusy runs fn again while the database is busy or locked by another
// jin process.
func retryBusy(fn func() error) error {
	var err error
	for attempt := 1; attempt <= busyRetries; attempt++ {
		if err = fn(); err == nil || !isBusy(err) {
			return err
		}
		time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
	}
	return err
}

func isBusy(err error) bool {
	text := err.Error()
	return strings.Contains(text, "SQLITE_BUSY") || strings.Contains(text, "SQLITE_LOCKED") ||
		strings.Contains(text, "database is locked") || strings.Contains(text, "database table is locked")
}
