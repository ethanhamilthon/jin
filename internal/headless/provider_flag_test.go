package headless

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"jin/internal/store"
)

// twoProviders saves providers "a" and "b" (a is active) and counts the
// requests each server gets.
func twoProviders(t *testing.T, h *harness) map[string]int {
	t.Helper()
	hits := map[string]int{}
	for _, id := range []string{"a", "b"} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			hits[id]++
			if r.URL.Path == "/models" {
				_, _ = w.Write([]byte(`{"data":[{"id":"model-` + id + `"}]}`))
				return
			}
			sse(w, answerChunk)
		}))
		t.Cleanup(server.Close)
		base := server.URL
		if id == "b" {
			base = strings.Replace(base, "http://", "http://user:secret@", 1)
		}
		if err := h.db.AddProvider(store.ProviderEntry{ID: id, Name: id, BaseURL: base, APIKey: "k"}); err != nil {
			t.Fatal(err)
		}
	}
	h.env = map[string]string{"JIN_MODEL": "m"}
	if err := h.db.SetActiveProvider("a"); err != nil {
		t.Fatal(err)
	}
	return hits
}

func TestProviderFlagBeatsSessionAndActive(t *testing.T) {
	h := newHarness(t, nil)
	hits := twoProviders(t, h)
	if code := h.run(t, "-p", "--provider", "b", "--format", "json", "first"); code != 0 {
		t.Fatalf("first: %d %s", code, h.errOut.String())
	}
	first := strings.SplitN(h.out.String(), "\n", 2)[0]
	for _, want := range []string{`"provider":"b"`, `"endpoint":"http://127.0.0.1:`} {
		if !strings.Contains(first, want) {
			t.Errorf("session record %s lacks %s", first, want)
		}
	}
	if strings.Contains(first, "secret") {
		t.Errorf("credentials in the session record: %s", first)
	}
	if code := h.run(t, "-p", "-c", "second"); code != 0 || hits["b"] != 2 || hits["a"] != 0 {
		t.Fatalf("session provider: code %d hits %v", code, hits)
	}
	if code := h.run(t, "-p", "-c", "--provider", "a", "third"); code != 0 || hits["a"] != 1 {
		t.Fatalf("flag over session: code %d hits %v", code, hits)
	}
	list, _ := h.db.ListByPath(h.dir)
	if list[0].Provider != "a" {
		t.Errorf("recorded provider %q, want a", list[0].Provider)
	}
}

func TestUnknownProviderListsKnownOnes(t *testing.T) {
	h := newHarness(t, nil)
	twoProviders(t, h)
	for _, args := range [][]string{{"-p", "--provider", "zzz", "hi"}, {"models", "--provider", "zzz"}} {
		if code := h.run(t, args...); code != exitError || !strings.Contains(h.errOut.String(), `unknown provider "zzz" (known: a, b)`) {
			t.Errorf("%v: code %d stderr %q", args, code, h.errOut.String())
		}
	}
}

func TestModelsProviderFlagListsThatProviderWithoutTouchingTheCache(t *testing.T) {
	h := newHarness(t, nil)
	hits := twoProviders(t, h)
	if code := h.run(t, "models", "--provider", "b"); code != 0 || h.out.String() != "model-b\t\t\t\t\n" || hits["b"] != 1 {
		t.Fatalf("code %d out %q hits %v err %q", code, h.out.String(), hits, h.errOut.String())
	}
	if ids, _ := h.db.LoadModelsCache(); len(ids) != 0 {
		t.Errorf("cache = %v, want untouched", ids)
	}
	if code := h.run(t, "refresh-models", "--provider", "b"); code != exitError {
		t.Errorf("refresh-models accepted --provider: %d", code)
	}
}

func TestCwdFlagSetsSessionPathAndProjectFiles(t *testing.T) {
	var body string
	var h *harness
	h = newHarness(t, captureBody(&h, &body))
	t.Chdir(t.TempDir())
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "AGENTS.md"), []byte("PROJECT-RULE-42"), 0o600); err != nil {
		t.Fatal(err)
	}
	h.dir = t.TempDir()
	if code := h.run(t, "-p", "--cwd", project, "hi"); code != 0 {
		t.Fatalf("code %d %s", code, h.errOut.String())
	}
	want, _ := filepath.EvalSymlinks(project)
	list, _ := h.db.ListByPath(project)
	if len(list) != 1 {
		t.Fatalf("no session stored for %s", project)
	}
	if got, _ := os.Getwd(); got != want {
		t.Errorf("process cwd %q, want %q", got, want)
	}
	if !strings.Contains(body, "PROJECT-RULE-42") {
		t.Error("AGENTS.md of --cwd was not read")
	}
	if code := h.run(t, "-p", "--cwd", filepath.Join(project, "AGENTS.md"), "hi"); code != exitError || !strings.Contains(h.errOut.String(), "is not a directory") {
		t.Errorf("file as cwd: code %d %q", code, h.errOut.String())
	}
}
