package provider

import (
	"errors"
	"net/url"
	"strings"
	"sync"
)

// Provider kinds. An empty Kind means KindOpenAI.
const (
	KindOpenAI    = "openai"
	KindAnthropic = "anthropic"
)

type Config struct {
	Kind    string
	BaseURL string
	APIKey  string
}

func (c Config) Ready() bool {
	return c.BaseURL != "" && c.APIKey != ""
}

func (c Config) Validate() error {
	parsed, err := url.Parse(c.BaseURL)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") ||
		parsed.Hostname() == "" || parsed.User != nil || strings.ContainsAny(c.BaseURL, "?#") {
		return errors.New("base URL must be a valid http(s) URL")
	}
	if strings.TrimSpace(c.APIKey) == "" {
		return errors.New("API key is required")
	}
	return nil
}

func NormalizeBaseURL(raw string) string {
	return strings.TrimRight(strings.TrimSpace(raw), "/")
}

type Client struct {
	mu         sync.RWMutex
	cfg        Config
	noThinking map[string]bool
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, noThinking: make(map[string]bool)}
}

func (c *Client) Configure(cfg Config) {
	c.mu.Lock()
	c.cfg = cfg
	c.mu.Unlock()
}

func (c *Client) Config() Config {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.cfg
}

func (c *Client) modelSupportsThinking(model string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.noThinking == nil {
		return true
	}
	return !c.noThinking[model]
}

func (c *Client) disableThinking(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.noThinking == nil {
		c.noThinking = make(map[string]bool)
	}
	c.noThinking[model] = true
}
