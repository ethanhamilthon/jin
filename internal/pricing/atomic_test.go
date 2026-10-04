package pricing

import (
	"bytes"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

func TestWriteAtomic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	data := []byte(`{"hello": "world"}`)
	if err := writeAtomic(path, data); err != nil {
		t.Fatalf("writeAtomic failed: %v", err)
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(got, data) {
		t.Fatalf("got %q, want %q, err: %v", got, data, err)
	}
}

func TestWriteAtomicConcurrent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cache.json")
	const workers = 10
	var wg sync.WaitGroup
	wg.Add(workers)

	for i := 0; i < workers; i++ {
		payload := bytes.Repeat([]byte{byte('a' + i)}, 1024)
		go func(data []byte) {
			defer wg.Done()
			_ = writeAtomic(path, data)
		}(payload)
	}
	wg.Wait()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read after concurrent writes failed: %v", err)
	}
	if len(data) != 1024 {
		t.Fatalf("corrupted file length: %d", len(data))
	}
	first := data[0]
	for _, b := range data {
		if b != first {
			t.Fatalf("torn write detected in atomic file")
		}
	}
}
