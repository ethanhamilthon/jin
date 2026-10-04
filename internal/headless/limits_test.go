package headless

import (
	"io"
	"net/http"
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
