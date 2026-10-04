package provider

import "strings"

// anthropicOptions are the optional request features a model may reject.
type anthropicOptions struct {
	thinking bool
	cache    bool
}

// featureKey scopes learned rejections to one endpoint, kind and model.
type featureKey struct {
	baseURL string
	kind    string
	model   string
}

type anthropicFeatures struct {
	noThinking bool
	noCache    bool
}

func (c *Client) featureKey(model string) featureKey {
	cfg := c.Config()
	return featureKey{baseURL: cfg.BaseURL, kind: cfg.Kind, model: model}
}

func (c *Client) knownFeatures(key featureKey) anthropicFeatures {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.features[key]
}

func (c *Client) updateFeatures(key featureKey, update func(*anthropicFeatures)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.features == nil {
		c.features = make(map[featureKey]anthropicFeatures)
	}
	features := c.features[key]
	update(&features)
	c.features[key] = features
}

func mentionsThinking(text string) bool {
	text = strings.ToLower(text)
	for _, word := range []string{"thinking", "effort", "output_config"} {
		if strings.Contains(text, word) {
			return true
		}
	}
	return false
}
