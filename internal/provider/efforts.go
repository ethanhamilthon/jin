package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
)

const invalidEffort = "__jin_invalid_effort__"

var DefaultEfforts = []string{"low", "medium", "high"}

var effortList = regexp.MustCompile(`(?i)\b(?:valid\s+(?:levels|values)|allowed\s+(?:levels|values)|one\s+of)\s*(?:are|is)?\s*:?\s*(["']?[a-z][a-z0-9_-]*["']?(?:\s*(?:,|\||\bor\b)\s*(?:or\s+)?["']?[a-z][a-z0-9_-]*["']?)*)`)
var effortName = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9_-]*`)

// Efforts probes the provider with an invalid reasoning_effort and parses the
// valid levels out of the rejection message. Nil means only the default exists.
func (c *Client) Efforts(ctx context.Context, model string) ([]string, error) {
	if c.Config().Kind == KindAnthropic {
		return DefaultEfforts, nil
	}
	payload, err := json.Marshal(struct {
		Model     string    `json:"model"`
		Messages  []Message `json:"messages"`
		Effort    string    `json:"reasoning_effort"`
		MaxTokens int       `json:"max_tokens"`
	}{model, []Message{{Role: "user", Content: "Hi"}}, invalidEffort, 1})
	if err != nil {
		return nil, err
	}
	status, body, err := c.request(ctx, http.MethodPost, "/chat/completions", payload)
	if err == nil {
		return nil, nil
	}
	if status != http.StatusBadRequest && status != http.StatusUnprocessableEntity {
		return DefaultEfforts, nil
	}
	var response struct {
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	message := string(body)
	if json.Unmarshal(body, &response) == nil && response.Error.Message != "" {
		message = response.Error.Message
	}
	if !strings.Contains(message, invalidEffort) && !strings.Contains(message, "reasoning_effort") {
		return DefaultEfforts, nil
	}
	match := effortList.FindStringSubmatch(message)
	if match == nil {
		return DefaultEfforts, nil
	}
	var levels []string
	for _, level := range effortName.FindAllString(match[1], -1) {
		if !strings.EqualFold(level, "or") {
			levels = append(levels, level)
		}
	}
	if len(levels) == 0 {
		return DefaultEfforts, nil
	}
	return levels, nil
}
