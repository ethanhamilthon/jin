package tools

import (
	"errors"
	"os"
	"sync"
	"time"
)

// Seen remembers how each file looked when the agent last read or wrote it,
// so that edit and write never overwrite a change made by someone else in
// the meantime. A nil *Seen checks nothing.
type Seen struct {
	mu    sync.Mutex
	files map[string]fileStamp
	// names maps each spelling the agent used to the identity of its file,
	// so a deleted file is still recognised.
	names map[string]string
}

type fileStamp struct {
	mod  time.Time
	size int64
}

func NewSeen() *Seen { return &Seen{files: map[string]fileStamp{}, names: map[string]string{}} }

// Remember records the file as the agent knows it now.
func (s *Seen) Remember(path string) {
	if s == nil {
		return
	}
	info, err := os.Stat(path)
	key := fileIdentity(path)
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		delete(s.files, key)
		return
	}
	s.names[path] = key
	s.files[key] = fileStamp{mod: info.ModTime(), size: info.Size()}
}

// Check fails when the file changed on disk since the agent last saw it.
// A file the agent never read, or one that does not exist, passes.
func (s *Seen) Check(path string) error {
	if s == nil {
		return nil
	}
	key := fileIdentity(path)
	s.mu.Lock()
	named, hasName := s.names[path]
	if hasName && key != named {
		s.mu.Unlock()
		return errors.New(path + " now resolves to a different file than the one you last read; read it again before writing it")
	}
	stamp, known := s.files[key]
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
		return errors.New(path + " changed on disk since you last read it (by the user or by a command such as sed -i, a formatter or git); read it again, then retry the change")
	}
	return nil
}
