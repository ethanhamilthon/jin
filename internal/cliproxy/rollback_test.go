package cliproxy

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestFailedReplacementRestoresThePreviousVersion(t *testing.T) {
	root := t.TempDir()
	installFixture(t, root, DefaultVersion)
	keys, err := loadKeys(root)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	current, err := startProcess(ctx, root, DefaultVersion, keys)
	if err != nil {
		t.Fatal(err)
	}
	b := &broker{root: root, keys: keys, proxy: current, wake: make(chan struct{}, 1)}
	defer func() { b.proxy.stop() }()
	b.install = func(_ context.Context, root, version string) error {
		if err := os.MkdirAll(filepath.Dir(binaryPath(root, version)), 0o700); err != nil {
			return err
		}
		exe, _ := os.Executable()
		return os.Symlink(exe, binaryPath(root, version))
	}
	err = b.update(ctx, "v8.0.25")
	if err == nil || !strings.Contains(err.Error(), "previous version was restored") {
		t.Fatalf("update=%v", err)
	}
	if Installed(root).Version != DefaultVersion {
		t.Fatal("failed replacement changed active version")
	}
	if _, err = b.proxy.management(ctx, "GET", "/auth-files", nil); err != nil {
		t.Fatalf("previous process not ready: %v", err)
	}
	b.mu.Lock()
	paused := b.updating
	b.mu.Unlock()
	if paused {
		t.Fatal("request leases remain blocked after recovery")
	}
}
