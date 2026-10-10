package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewer(t *testing.T) {
	cases := map[[2]string]bool{
		{"v0.6.2", "v0.6.1"}: true, {"v0.7", "v0.6.9"}: true, {"v0.6.1", "v0.6"}: true,
		{"v0.6", "v0.6.0"}: false, {"v0.6.1", "v0.6.1"}: false, {"v0.6.1", "v0.7"}: false,
		{"v1.0", "dev"}: false, {"garbage", "v0.1"}: false,
	}
	for in, want := range cases {
		if got := Newer(in[0], in[1]); got != want {
			t.Errorf("Newer(%s, %s) = %v", in[0], in[1], got)
		}
	}
}

func archiveWith(t *testing.T, body string) []byte {
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	_ = tw.WriteHeader(&tar.Header{Name: "jin", Mode: 0o755, Size: int64(len(body)), Typeflag: tar.TypeReg})
	_, _ = tw.Write([]byte(body))
	_ = tw.Close()
	_ = gz.Close()
	return buf.Bytes()
}

func TestLatestAndInstall(t *testing.T) {
	archive := archiveWith(t, "#!/bin/sh\necho new\n")
	sum := sha256.Sum256(archive)
	sums := hex.EncodeToString(sum[:]) + "  " + Archive() + "\n"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			http.Redirect(w, r, "/"+Repo+"/releases/tag/v9.9.9", http.StatusFound)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
			_, _ = w.Write([]byte(sums))
		case strings.HasSuffix(r.URL.Path, "/"+Archive()):
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	releaseBase = server.URL
	defer func() { releaseBase = "https://github.com" }()
	tag, err := Latest(t.Context())
	if err != nil || tag != "v9.9.9" {
		t.Fatalf("latest: %q %v", tag, err)
	}
	target := filepath.Join(t.TempDir(), "jin")
	_ = os.WriteFile(target, []byte("old"), 0o755)
	if err := Install(t.Context(), tag, target); err != nil {
		t.Fatal(err)
	}
	if data, _ := os.ReadFile(target); string(data) != "#!/bin/sh\necho new\n" {
		t.Fatalf("binary = %q", data)
	}
	sums = strings.Repeat("0", 64) + "  " + Archive() + "\n"
	if err := Install(t.Context(), tag, target); err == nil || !strings.Contains(err.Error(), "checksum") {
		t.Fatalf("bad checksum: %v", err)
	}
}

// A daemon that is busy keeps the old binary: the update refuses and says how
// to force it. The daemon is stopped first only with --force.
func TestUpdateStopsTheDaemonFirst(t *testing.T) {
	var stopped, started bool
	archive := archiveWith(t, "new")
	sum := sha256.Sum256(archive)
	sums := hex.EncodeToString(sum[:]) + "  " + Archive() + "\n"
	control := Control{
		Stop: func(_ context.Context, force bool, out io.Writer) error {
			if !force {
				return errors.New("the running daemon is busy; run jin update --force")
			}
			stopped = true
			fmt.Fprintln(out, "stopped the running daemon")
			return nil
		},
		Start: func(_ context.Context, out io.Writer) error {
			started = true
			_, _ = fmt.Fprintln(out, "started the daemon again")
			return nil
		},
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/releases/latest"):
			http.Redirect(w, r, "/"+Repo+"/releases/tag/v9.9.9", http.StatusFound)
		case strings.HasSuffix(r.URL.Path, "/checksums.txt"):
			_, _ = w.Write([]byte(sums))
		case strings.HasSuffix(r.URL.Path, "/"+Archive()):
			_, _ = w.Write(archive)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	releaseBase = server.URL
	defer func() { releaseBase = "https://github.com" }()
	var out, errOut strings.Builder
	if code := Main(t.Context(), nil, "v0.0.1", control, &out, &errOut); code != 1 {
		t.Fatalf("busy daemon: code %d, %s", code, errOut.String())
	}
	if stopped {
		t.Fatal("the daemon was stopped without force")
	}
	if !strings.Contains(errOut.String(), "--force") {
		t.Fatalf("the error does not name --force: %s", errOut.String())
	}
	old := executable
	defer func() { executable = old }()
	target := filepath.Join(t.TempDir(), "jin")
	_ = os.WriteFile(target, []byte("old"), 0o755)
	executable = func() (string, error) { return target, nil }
	out.Reset()
	errOut.Reset()
	if code := Main(t.Context(), []string{"--force"}, "v0.0.1", control, &out, &errOut); code != 0 {
		t.Fatalf("forced update: code %d, %s", code, errOut.String())
	}
	if !stopped || !started {
		t.Fatalf("stop=%t start=%t", stopped, started)
	}
	if data, _ := os.ReadFile(target); string(data) != "new" {
		t.Fatalf("binary was not replaced: %q", data)
	}
}
