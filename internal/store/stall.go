package store

import (
	"strconv"
	"time"
)

const keyStallTimeout = "provider.stall_timeout"

// parseStallTimeout reads the stream silence limit in seconds; 0 means the
// provider default.
func parseStallTimeout(value string) time.Duration {
	seconds, err := strconv.Atoi(value)
	if err != nil || seconds <= 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
