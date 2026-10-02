package headless

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"jin/internal/pricing"
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
	if h.out.String() != "a\t\t\t\tlow,medium,high\nb\t\t\t\t\nc\t\t\t\tlow,medium,high\n" {
		t.Fatalf("models text = %q", h.out.String())
	}
	h.run(t, "models", "--format", "json")
	if strings.TrimSpace(h.out.String()) != `[{"id":"a","efforts":["low","medium","high"]},{"id":"b"},{"id":"c","efforts":["low","medium","high"]}]` {
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

func TestModelsScopeAndAll(t *testing.T) {
	h := newHarness(t, nil)
	server := modelsServer(t)
	defer server.Close()
	h.env["JIN_BASE_URL"] = server.URL

	if code := h.run(t, "refresh-models"); code != 0 {
		t.Fatalf("refresh: %d", code)
	}

	if err := h.db.SaveScope([]string{"a", "c"}); err != nil {
		t.Fatal(err)
	}

	h.run(t, "models")
	if h.out.String() != "a\t\t\t\t\nc\t\t\t\t\n" {
		t.Fatalf("scoped text = %q", h.out.String())
	}

	h.run(t, "models", "--all")
	if h.out.String() != "a\t\t\t\t\nb\t\t\t\t\nc\t\t\t\t\n" {
		t.Fatalf("all text = %q", h.out.String())
	}

	if code := h.run(t, "refresh-models", "--all"); code != 1 || !strings.Contains(h.errOut.String(), "--all belongs to models") {
		t.Fatalf("refresh --all code %d err %q", code, h.errOut.String())
	}
}

func TestModelsPricingAndContext(t *testing.T) {
	h := newHarness(t, nil)
	server := modelsServer(t)
	defer server.Close()
	h.env["JIN_BASE_URL"] = server.URL

	loadPricing = func(context.Context) pricing.Table {
		return pricing.Table{
			"a": {InputCostPerToken: 1.25e-6, OutputCostPerToken: 5e-6, MaxInputTokens: 128000},
			"b": {MaxInputTokens: 64000},
		}
	}

	if code := h.run(t, "refresh-models", "--efforts"); code != 0 {
		t.Fatalf("refresh: %d", code)
	}

	h.run(t, "models")
	wantText := "a\t1.25\t5\t128000\tlow,medium,high\n" +
		"b\t\t\t64000\t\n" +
		"c\t\t\t\tlow,medium,high\n"
	if h.out.String() != wantText {
		t.Fatalf("models text =\n%q\nwant:\n%q", h.out.String(), wantText)
	}

	h.run(t, "models", "--format", "json")
	wantJSON := `[{"id":"a","input_per_mtok":1.25,"output_per_mtok":5,"context_window":128000,"efforts":["low","medium","high"]},{"id":"b","context_window":64000},{"id":"c","efforts":["low","medium","high"]}]`
	if strings.TrimSpace(h.out.String()) != wantJSON {
		t.Fatalf("models json =\n%q\nwant:\n%q", h.out.String(), wantJSON)
	}
}

func TestModelsEmptyJSON(t *testing.T) {
	var buf bytes.Buffer
	if err := printModels(&buf, nil, "json"); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Fatalf("nil entries json = %q", buf.String())
	}
	buf.Reset()
	if err := printModels(&buf, []modelEntry{}, "json"); err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(buf.String()) != "[]" {
		t.Fatalf("empty slice json = %q", buf.String())
	}
}
