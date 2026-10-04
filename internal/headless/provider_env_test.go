package headless

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// The flag beats the environment; without the flag the environment still
// replaces the saved provider.
func TestProviderFlagBeatsProviderEnvironment(t *testing.T) {
	h := newHarness(t, nil)
	hits := twoProviders(t, h)
	decoyHits := 0
	decoy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		decoyHits++
		if r.URL.Path == "/models" {
			_, _ = w.Write([]byte(`{"data":[{"id":"decoy"}]}`))
			return
		}
		sse(w, answerChunk)
	}))
	t.Cleanup(decoy.Close)
	h.env["JIN_BASE_URL"], h.env["JIN_API_KEY"], h.env["JIN_PROVIDER_KIND"] = decoy.URL, "env-key", "openai"

	for _, args := range [][]string{{"-p", "--no-session", "--provider", "b", "hi"}, {"models", "--provider", "b"}} {
		if code := h.run(t, args...); code != 0 {
			t.Fatalf("%v: code %d stderr %q", args, code, h.errOut.String())
		}
	}
	if hits["b"] != 2 || decoyHits != 0 {
		t.Fatalf("with the flag: hits %v, decoy %d", hits, decoyHits)
	}
	for _, args := range [][]string{{"-p", "--no-session", "hi"}, {"models"}} {
		if code := h.run(t, args...); code != 0 {
			t.Fatalf("%v: code %d stderr %q", args, code, h.errOut.String())
		}
	}
	if decoyHits != 2 || hits["b"] != 2 {
		t.Fatalf("without the flag: hits %v, decoy %d", hits, decoyHits)
	}
}
