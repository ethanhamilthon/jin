package search

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"sync"
)

type Result struct {
	Title, URL, Snippet string
}

type Searcher interface {
	Search(ctx context.Context, query string, limit int) ([]Result, error)
}

type Backend string

const (
	Duck   Backend = "duck"
	Brave  Backend = "brave"
	Tavily Backend = "tavily"
)

var Backends = []Backend{Duck, Brave, Tavily}

func (b Backend) NeedsKey() bool { return b != Duck }

type Config struct {
	Backend Backend
	Keys    map[Backend]string
}

func (c Config) Searcher() (Searcher, error) {
	key := c.Keys[c.Backend]
	if c.Backend.NeedsKey() && key == "" {
		return nil, errors.New(string(c.Backend) + " search needs an API key: press w in NORMAL mode")
	}
	switch c.Backend {
	case Brave:
		return brave{key: key}, nil
	case Tavily:
		return tavily{key: key}, nil
	default:
		return duck{}, nil
	}
}

// Settings is shared between the UI, which edits it, and the websearch tool.
type Settings struct {
	mu  sync.RWMutex
	cfg Config
}

func (s *Settings) Set(cfg Config) {
	s.mu.Lock()
	s.cfg = cfg
	s.mu.Unlock()
}

func (s *Settings) Get() Config {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.cfg
}

func Format(results []Result) string {
	if len(results) == 0 {
		return "No results"
	}
	var b strings.Builder
	for i, r := range results {
		b.WriteString(strconv.Itoa(i+1) + ". " + r.Title + "\n   " + r.URL + "\n")
		if r.Snippet != "" {
			b.WriteString("   " + r.Snippet + "\n")
		}
	}
	return b.String()
}
