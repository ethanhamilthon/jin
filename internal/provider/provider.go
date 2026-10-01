package provider

import (
	"errors"
	"net/url"
	"strings"
	"sync"
)

type Config struct {
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

type Message struct {
	Role             string     `json:"role"`
	Content          string     `json:"content,omitempty"`
	ReasoningContent string     `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string     `json:"tool_call_id,omitempty"`
}

type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

type Client struct {
	mu  sync.RWMutex
	cfg Config
}

func NewClient(cfg Config) *Client {
	return &Client{cfg: cfg}
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
