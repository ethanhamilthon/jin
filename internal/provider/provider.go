package provider

import (
	"errors"
	"net/url"
	"strings"
	"sync"
	"time"
)

// Provider kinds. An empty Kind means KindOpenAI.
const (
	KindOpenAI    = "openai"
	KindResponses = "responses"
	KindAnthropic = "anthropic"
)

type Config struct {
	Kind     string
	BaseURL  string
	APIKey   string
	Managed  bool
	Disabled bool
}

func (c Config) Ready() bool {
	return !c.Disabled && (c.Managed || (c.BaseURL != "" && c.APIKey != ""))
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
	mu          sync.RWMutex
	cfg         Config
	features    map[featureKey]anthropicFeatures
	stall       time.Duration
	debug       *debugState
	target      RequestTarget
	modelPrefix string
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg, debug: newDebugState()}
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
