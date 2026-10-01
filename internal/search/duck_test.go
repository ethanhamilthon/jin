package search

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

const duckPage = `<div class="results">
<div class="result result--ad"><a class="result__a" href="https://ads.example">Ad</a></div>
<div class="result results_links web-result">
  <h2 class="result__title"><a class="result__a" href="//duckduckgo.com/l/?uddg=https%3A%2F%2Fgo.dev%2Fdoc%2F&amp;rut=x">Go <b>docs</b></a></h2>
  <a class="result__snippet">The   Go
  documentation.</a>
</div>
<div class="result web-result"><a class="result__a" href="https://example.com/b">Second</a></div>
</div>`

func TestParseDuckSkipsAdsAndUnwrapsRedirects(t *testing.T) {
	doc, err := html.Parse(strings.NewReader(duckPage))
	if err != nil {
		t.Fatal(err)
	}
	got := parseDuck(doc, 10)
	want := []Result{
		{Title: "Go docs", URL: "https://go.dev/doc/", Snippet: "The Go documentation."},
		{Title: "Second", URL: "https://example.com/b"},
	}
	if len(got) != len(want) {
		t.Fatalf("got %d results: %+v", len(got), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("result %d = %+v, want %+v", i, got[i], want[i])
		}
	}
	if limited := parseDuck(doc, 1); len(limited) != 1 {
		t.Errorf("limit 1 returned %d results", len(limited))
	}
}
