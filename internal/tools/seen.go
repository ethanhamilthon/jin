package tools

import (
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Seen remembers how each file looked when the agent last read or wrote it,
// so that edit and write never overwrite a change made by someone else in
// the meantime. A nil *Seen checks nothing.
type Seen struct {
	mu    sync.Mutex
	files map[string]fileStamp
}

type fileStamp struct {
	mod  time.Time
	size int64
}

func NewSeen() *Seen { return &Seen{files: map[string]fileStamp{}} }

func seenKey(path string) string {
	if abs, err := filepath.Abs(path); err == nil {
		return abs
	}
	return path
}

// Remember records the file as the agent knows it now.
func (s *Seen) Remember(path string) {
	if s == nil {
		return
	}
	info, err := os.Stat(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		delete(s.files, seenKey(path))
		return
	}
	s.files[seenKey(path)] = fileStamp{mod: info.ModTime(), size: info.Size()}
}

// Check fails when the file changed on disk since the agent last saw it.
// A file the agent never read, or one that does not exist, passes.
func (s *Seen) Check(path string) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	stamp, known := s.files[seenKey(path)]
	s.mu.Unlock()
	if !known {
		return nil
	}
	info, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return errors.New(path + " was deleted since you last read it; check with the user or read the directory again before writing it")
	}
	if err != nil {
		return err
	}
	if !info.ModTime().Equal(stamp.mod) || info.Size() != stamp.size {
		return errors.New(path + " changed on disk since you last read it (maybe the user edited it); read it again, then retry the change")
	}
	return nil
}
