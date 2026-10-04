package headless

import (
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

func TestTimeoutCoversStdin(t *testing.T) {
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { sse(w, answerChunk) })
	reader, writer := io.Pipe()
	defer writer.Close()
	started := time.Now()
	code := Run(t.Context(), []string{"-p", "--timeout", "300ms"}, h.db, h.dir, h.signals, ioSet{
		in: reader, piped: true, out: &h.out, err: &h.errOut,
		getenv: func(k string) string { return h.env[k] },
	})
	if code != exitError || !strings.Contains(h.errOut.String(), "timed out after 300ms") {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	if time.Since(started) > 5*time.Second {
		t.Fatal("run did not stop at the timeout")
	}
}

func TestTimeoutDuringPromptCommandSendsNothing(t *testing.T) {
	requests := 0
	h := newHarness(t, func(w http.ResponseWriter, r *http.Request) { requests++; sse(w, answerChunk) })
	h.dir = t.TempDir()
	root := os.Getenv("HOME") + "/.jin-dev"
	if err := os.MkdirAll(root, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(root+"/system-prompt.md", []byte("# system\n\n{{sleep 5}}\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	started := time.Now()
	code := h.run(t, "-p", "--timeout", "300ms", "hi")
	if code != exitError || !strings.Contains(h.errOut.String(), "timed out after 300ms") {
		t.Fatalf("code %d, stderr %q", code, h.errOut.String())
	}
	if requests != 0 || time.Since(started) > 4*time.Second {
		t.Fatalf("requests %d, took %v", requests, time.Since(started))
	}
}
