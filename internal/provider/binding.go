package provider

func (c *Client) Bind(other *Client) {
	other.mu.RLock()
	cfg, target, prefix := other.cfg, other.target, other.modelPrefix
	other.mu.RUnlock()
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg, c.target, c.modelPrefix = cfg, target, prefix
}
