package sysprompt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func serveLatest(t *testing.T, texts map[string]string) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/"), ".md")
		text, ok := texts[name]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write([]byte(text))
	}))
	old := LatestBase
	LatestBase = server.URL + "/"
	t.Cleanup(func() { LatestBase = old; server.Close() })
}

func TestLatestFetchesTheNamedSections(t *testing.T) {
	serveLatest(t, map[string]string{"system": " new system \n", "compact": "new compact"})
	got, err := Latest(context.Background(), SectionSystem, SectionCompact)
	if err != nil || got[SectionSystem] != "new system" || got[SectionCompact] != "new compact" {
		t.Fatalf("got %v, %v", got, err)
	}
}

func TestLatestRefusesUnknownNamesAndMissingFiles(t *testing.T) {
	serveLatest(t, map[string]string{"system": "x"})
	if _, err := Latest(context.Background(), "../secret"); err == nil {
		t.Error("an unknown name must not be fetched")
	}
	if _, err := Latest(context.Background(), SectionHandoff); err == nil || !strings.Contains(err.Error(), "404") {
		t.Errorf("missing file error = %v", err)
	}
}

func TestResetReplacesOnlyTheNamedSections(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	mine := Sections{System: "my system", Compact: "my compact", Handoff: "my handoff"}
	path, err := Path()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := Create(); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(mine.Render()), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := Reset(map[string]string{SectionCompact: "latest compact"}); err != nil {
		t.Fatal(err)
	}
	got, err := Load()
	if err != nil || got.System != "my system" || got.Compact != "latest compact" || got.Handoff != "my handoff" {
		t.Fatalf("got %+v, %v", got, err)
	}
	if _, err := Reset(map[string]string{"nope": "x"}); err == nil {
		t.Error("an unknown section must be refused")
	}
}
