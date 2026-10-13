package provider

func (c *Client) Bind(other *Client) {
	other.mu.RLock()
	cfg, target, prefix := other.cfg, other.target, other.modelPrefix
	other.mu.RUnlock()
	other.protocolMu.Lock()
	protocols := other.protocols
	other.protocolMu.Unlock()
	if protocols == nil {
		c.protocolMu.Lock()
		protocols = c.protocols
		c.protocolMu.Unlock()
	}
	if protocols != nil {
		protocols.mu.RLock()
		load, save := protocols.load, protocols.save
		protocols.mu.RUnlock()
		if load != nil || save != nil {
			local := &protocolState{selected: map[string]string{}, load: load, save: save}
			protocols = local
		}
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cfg, c.target, c.modelPrefix = cfg, target, prefix
	c.protocolMu.Lock()
	c.protocols = protocols
	c.protocolMu.Unlock()
}
