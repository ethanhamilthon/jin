package search

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"golang.org/x/net/html"
)

const duckEndpoint = "https://html.duckduckgo.com/html/"

type duck struct{}

// Search posts the form the way the HTML frontend does; DuckDuckGo answers
// bare GET requests with a bot challenge.
func (duck) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	form := url.Values{"q": {query}, "b": {""}}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, duckEndpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "text/html")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Referer", "https://html.duckduckgo.com/")
	req.Header.Set("Origin", "https://html.duckduckgo.com")
	body, err := fetch(req, "")
	if err != nil {
		return nil, err
	}
	doc, err := html.Parse(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	results := parseDuck(doc, limit)
	if len(results) == 0 && bytes.Contains(body, []byte("anomaly")) {
		return nil, errors.New("DuckDuckGo blocked the request with a bot challenge; try again later or switch backend with w")
	}
	return results, nil
}

func parseDuck(doc *html.Node, limit int) []Result {
	var results []Result
	for node := range doc.Descendants() {
		if len(results) >= limit {
			break
		}
		if node.Type != html.ElementNode || !hasClass(node, "result") || hasClass(node, "result--ad") {
			continue
		}
		var r Result
		for child := range node.Descendants() {
			switch {
			case hasClass(child, "result__a") && r.Title == "":
				r.Title, r.URL = textOf(child), duckTarget(attr(child, "href"))
			case hasClass(child, "result__snippet") && r.Snippet == "":
				r.Snippet = textOf(child)
			}
		}
		if r.Title != "" && r.URL != "" {
			results = append(results, r)
		}
	}
	return results
}

func duckTarget(href string) string {
	parsed, err := url.Parse(href)
	if err != nil {
		return href
	}
	if target := parsed.Query().Get("uddg"); target != "" {
		return target
	}
	if parsed.Scheme == "" {
		parsed.Scheme = "https"
	}
	return parsed.String()
}
