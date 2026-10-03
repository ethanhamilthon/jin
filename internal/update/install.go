package update

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"
)

const maxArchive = 128 << 20

// Archive is the release file for this system, e.g. jin_darwin_arm64.tar.gz.
func Archive() string { return "jin_" + runtime.GOOS + "_" + runtime.GOARCH + ".tar.gz" }

// Install downloads tag, checks its SHA-256 against checksums.txt and
// replaces the binary at target. The new file is written next to target and
// renamed over it, so a failure leaves the old binary in place.
func Install(ctx context.Context, tag, target string) error {
	base := releaseBase + "/" + Repo + "/releases/download/" + tag + "/"
	sums, err := download(ctx, base+"checksums.txt")
	if err != nil {
		return err
	}
	archive, err := download(ctx, base+Archive())
	if err != nil {
		return err
	}
	if err := verify(sums, Archive(), archive); err != nil {
		return err
	}
	binary, err := extract(archive)
	if err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(target), ".jin-update-*")
	if err != nil {
		return fmt.Errorf("cannot write next to %s: %w", target, err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(binary); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), target)
}

func download(ctx context.Context, url string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download %s: HTTP %d", url, resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, maxArchive))
}
