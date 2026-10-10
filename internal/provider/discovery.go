package provider

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"sort"
	"strings"
)

func (c *Client) Models(ctx context.Context) ([]string, error) {
	path := "/models"
	if c.Config().Kind == KindAnthropic {
		path = "/v1/models"
	}
	_, body, err := c.request(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var response struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &response); err != nil || response.Data == nil {
		return nil, errors.New("invalid models response")
	}
	ids := make([]string, 0, len(response.Data))
	seen := make(map[string]bool, len(response.Data))
	c.mu.RLock()
	prefix := c.modelPrefix
	c.mu.RUnlock()
	for _, model := range response.Data {
		if prefix != "" && !strings.HasPrefix(model.ID, prefix) {
			continue
		}
		if strings.TrimSpace(model.ID) != "" && !seen[model.ID] {
			ids = append(ids, model.ID)
			seen[model.ID] = true
		}
	}
	sort.Strings(ids)
	return ids, nil
}
