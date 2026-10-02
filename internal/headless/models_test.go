package headless

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func modelsServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/models" {
			_, _ = w.Write([]byte(`{"data":[{"id":"b"},{"id":"a"},{"id":"c"}]}`))
			return
		}
		body := new(bytes.Buffer)
		body.ReadFrom(r.Body)
		switch {
		case strings.Contains(body.String(), `"model":"a"`):
			http.Error(w, `{"error":{"message":"invalid reasoning_effort, valid levels: low, medium, high"}}`, 400)
		case strings.Contains(body.String(), `"model":"b"`):
			_, _ = w.Write([]byte(`{}`))
		default:
			http.Error(w, "oops", 500)
		}
	}))
}

func TestRefreshAndListModels(t *testing.T) {
	h := newHarness(t, nil)
	server := modelsServer(t)
	defer server.Close()
	h.env["JIN_BASE_URL"] = server.URL
	if code := h.run(t, "models", "--format", "json"); code != 0 || !strings.Contains(h.out.String(), `[{"id":"a"},{"id":"b"},{"id":"c"}]`) {
		t.Fatalf("models fetch: %d %q %q", code, h.out.String(), h.errOut.String())
	}
	if code := h.run(t, "refresh-models", "--efforts"); code != 0 || !strings.Contains(h.errOut.String(), "3 models") {
		t.Fatalf("refresh: %d %q", code, h.errOut.String())
	}
	h.run(t, "models")
	if h.out.String() != "a\tlow,medium,high\nb\nc\n" {
		t.Fatalf("models text = %q", h.out.String())
	}
	h.run(t, "models", "--format", "json")
	if strings.TrimSpace(h.out.String()) != `[{"id":"a","efforts":["low","medium","high"]},{"id":"b"},{"id":"c"}]` {
		t.Fatalf("models json = %q", h.out.String())
	}
}

func TestModelsNeedProvider(t *testing.T) {
	h := newHarness(t, nil)
	h.env = map[string]string{}
	if code := h.run(t, "models"); code != 1 || !strings.Contains(h.errOut.String(), "JIN_BASE_URL") {
		t.Fatalf("code %d %q", code, h.errOut.String())
	}
}
