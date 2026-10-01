package search

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
)

type tavily struct{ key string }

func (t tavily) Search(ctx context.Context, query string, limit int) ([]Result, error) {
	payload, err := json.Marshal(struct {
		Query      string `json:"query"`
		MaxResults int    `json:"max_results"`
	}{query, limit})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://api.tavily.com/search", bytes.NewReader(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+t.key)
	body, err := fetch(req, t.key)
	if err != nil {
		return nil, err
	}
	var response struct {
		Results []struct {
			Title   string `json:"title"`
			URL     string `json:"url"`
			Content string `json:"content"`
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, err
	}
	var results []Result
	for _, r := range response.Results {
		results = append(results, Result{Title: r.Title, URL: r.URL, Snippet: r.Content})
	}
	return results[:min(len(results), limit)], nil
}
