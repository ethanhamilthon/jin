package tools

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"sync"
	"testing"
)

func TestSeenDetectsChangeWithSameSizeAndTime(t *testing.T) {
	path := writeTempFile(t, "one")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	seen := NewSeen()
	readThrough(t, seen, path)
	if err := os.WriteFile(path, []byte("two"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}
	if err := seen.Check(path); err == nil {
		t.Fatal("content change must be detected")
	}
}

func TestAtomicWriteRefusesContentConflict(t *testing.T) {
	path := writeTempFile(t, "original")
	check := unchangedContent(path, "original", true)
	if err := os.WriteFile(path, []byte("user change"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := writeFileAtomic(path, []byte("agent change"), 0o644, check); err == nil {
		t.Fatal("expected conflict")
	}
	data, _ := os.ReadFile(path)
	if string(data) != "user change" {
		t.Fatalf("lost user change: %s", data)
	}
}

func TestConcurrentSessionsKeepBothEdits(t *testing.T) {
	path := writeTempFile(t, "one two")
	alias := path + "-alias"
	if err := os.Symlink(path, alias); err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i, old := range []string{"one", "two"} {
		wg.Go(func() {
			<-start
			target := path
			if i == 1 {
				target = alias
			}
			args, _ := json.Marshal(map[string]string{"path": target, "old_string": old, "new_string": strings.ToUpper(old)})
			if _, err := NewEdit().Run(context.Background(), string(args)); err != nil {
				t.Error(err)
			}
		})
	}
	close(start)
	wg.Wait()
	data, _ := os.ReadFile(path)
	if string(data) != "ONE TWO" {
		t.Fatalf("lost concurrent edit: %s", data)
	}
}
