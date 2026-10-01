package tools

import (
	"bytes"
	"sync"
)

type boundedOutput struct {
	mu        sync.Mutex
	buffer    bytes.Buffer
	truncated bool
}

func (out *boundedOutput) Write(p []byte) (int, error) {
	out.mu.Lock()
	defer out.mu.Unlock()
	n := len(p)
	remaining := maxBashOutput - out.buffer.Len()
	if remaining > 0 {
		if n < remaining {
			remaining = n
		}
		out.buffer.Write(p[:remaining])
	}
	if n > remaining {
		out.truncated = true
	}
	return n, nil
}
