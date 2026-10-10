package tools

import "sync"

type fileLock struct {
	mu    sync.RWMutex
	users int
}

var fileLocks = struct {
	sync.Mutex
	files map[string]*fileLock
}{files: make(map[string]*fileLock)}

// lockFile serializes changes across sessions while allowing concurrent reads.
func lockFile(path string, reading ...bool) func() {
	read := len(reading) > 0 && reading[0]
	key := fileIdentity(path)
	fileLocks.Lock()
	lock := fileLocks.files[key]
	if lock == nil {
		lock = &fileLock{}
		fileLocks.files[key] = lock
	}
	lock.users++
	fileLocks.Unlock()
	if read {
		lock.mu.RLock()
	} else {
		lock.mu.Lock()
	}
	return func() {
		if read {
			lock.mu.RUnlock()
		} else {
			lock.mu.Unlock()
		}
		fileLocks.Lock()
		lock.users--
		if lock.users == 0 {
			delete(fileLocks.files, key)
		}
		fileLocks.Unlock()
	}
}
