package search

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

type brave struct{ key string }

func (b brave) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	endpoint := "https://api.search.brave.com/res/v1/web/search?q=" + url.QueryEscape(query) + "&count=" + strconv.Itoa(limit)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Subscription-Token", b.key)
	body, err := fetch(req, b.key)
	if err != nil {
		return nil, err
	}
	var response struct {
		Web struct {
			Results []struct {
				Title       string `json:"title"`
				URL         string `json:"url"`
				Description string `json:"description"`
			} `json:"results"`
		} `json:"web"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	var results []Result
	for _, r := range response.Web.Results {
		results = append(results, Result{Title: stripTags(r.Title), URL: r.URL, Snippet: stripTags(r.Description)})
	}
	return results[:min(len(results), limit)], nil
}
