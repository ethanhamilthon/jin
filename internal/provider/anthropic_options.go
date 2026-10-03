package provider

// anthropicOptions are the optional request features a model may reject.
type anthropicOptions struct {
	thinking bool
	cache    bool
}

func (c *Client) modelSupportsCache(model string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return !c.noCache[model]
}

func (c *Client) disableCache(model string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.noCache == nil {
		c.noCache = make(map[string]bool)
	}
	c.noCache[model] = true
}
