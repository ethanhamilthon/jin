package cliproxy

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestUpdateWaitsForRequestsAndPreservesAuth(t *testing.T) {
	root := t.TempDir()
	installFixture(t, root, DefaultVersion)
	keys, err := loadKeys(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	p, err := startProcess(ctx, root, DefaultVersion, keys)
	if err != nil {
		t.Fatal(err)
	}
	b := &broker{root: root, keys: keys, proxy: p, active: 1, wake: make(chan struct{}, 1)}
	defer func() { b.proxy.stop() }()
	b.install = func(_ context.Context, root, version string) error {
		if err := os.MkdirAll(filepath.Dir(binaryPath(root, version)), 0o700); err != nil {
			return err
		}
		exe, _ := os.Executable()
		return os.Symlink(exe, binaryPath(root, version))
	}
	auth := filepath.Join(root, "auth", "keep.json")
	if err := os.WriteFile(auth, []byte(`{"fixture":"preserve"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- b.update(ctx, "v8.0.24") }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		b.mu.Lock()
		paused := b.updating
		b.mu.Unlock()
		if paused {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("update did not reach idle wait")
		}
		time.Sleep(10 * time.Millisecond)
	}
	if Installed(root).Version != DefaultVersion {
		t.Fatal("changed version with active request")
	}
	b.mu.Lock()
	b.active = 0
	b.mu.Unlock()
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	installed := Installed(root)
	if installed.Version != "v8.0.24" || installed.Previous != DefaultVersion {
		t.Fatalf("installation=%+v", installed)
	}
	if data, _ := os.ReadFile(auth); string(data) != `{"fixture":"preserve"}` {
		t.Fatal("auth was overwritten")
	}
}
