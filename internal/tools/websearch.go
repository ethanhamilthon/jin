package tools

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"jin/internal/search"
)

const defaultSearchLimit = 8
const maxSearchLimit = 20

const webSearchSchema = `{"type":"function","function":{"name":"websearch","description":"Search the web and return titles, URLs, and snippets","parameters":{"type":"object","properties":{"query":{"type":"string","description":"Search query"},"limit":{"type":"integer","description":"Maximum number of results, 1-20. Defaults to 8."}},"required":["query"],"additionalProperties":false}}}`

type WebSearch struct {
	settings *search.Settings
}

func NewWebSearch(settings *search.Settings) WebSearch { return WebSearch{settings: settings} }

func (WebSearch) Name() string { return "websearch" }

func (WebSearch) Schema() json.RawMessage { return json.RawMessage(webSearchSchema) }

func (w WebSearch) Summary(argumentsJSON string) (string, bool) {
	query, _, ok := parseSearchArgs(argumentsJSON)
	if !ok {
		return "", false
	}
	return query + " (" + string(w.settings.Get().Backend) + ")", true
}

func (w WebSearch) Run(ctx context.Context, argumentsJSON string) (string, error) {
	query, limit, ok := parseSearchArgs(argumentsJSON)
	if !ok {
		return "", errors.New("invalid websearch tool arguments")
	}
	searcher, err := w.settings.Get().Searcher()
	if err != nil {
		return "", err
	}
	results, err := searcher.Search(ctx, query, limit)
	if err != nil {
		return "", err
	}
	return search.Format(results), nil
}

func parseSearchArgs(argumentsJSON string) (string, int, bool) {
	var args struct {
		Query string `json:"query"`
		Limit int    `json:"limit"`
	}
	if json.Unmarshal([]byte(argumentsJSON), &args) != nil || strings.TrimSpace(args.Query) == "" {
		return "", 0, false
	}
	limit := args.Limit
	if limit <= 0 {
		limit = defaultSearchLimit
	}
	return strings.TrimSpace(args.Query), min(limit, maxSearchLimit), true
}
