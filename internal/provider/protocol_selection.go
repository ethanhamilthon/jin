package provider

import "sync"

type protocolState struct {
	mu       sync.RWMutex
	selected map[string]string
	load     func(string) (string, error)
	save     func(string, string) error
}

func (c *Client) protocolState() *protocolState {
	c.protocolMu.Lock()
	defer c.protocolMu.Unlock()
	if c.protocols == nil {
		c.protocols = &protocolState{selected: map[string]string{}}
	}
	return c.protocols
}

func (c *Client) SetProtocolStore(load func(model string) (string, error), save func(model, protocol string) error) {
	p := c.protocolState()
	p.mu.Lock()
	p.load, p.save = load, save
	p.mu.Unlock()
}

func validProtocol(protocol string) bool { return protocol == KindOpenAI || protocol == KindResponses }

func initialProtocol(cfg Config) string {
	if validProtocol(cfg.Protocol) {
		return cfg.Protocol
	}
	return KindResponses
}

func (c *Client) protocolFor(model string, cfg Config) string {
	p := c.protocolState()
	p.mu.RLock()
	load, selected := p.load, p.selected[model]
	p.mu.RUnlock()
	if selected != "" {
		return selected
	}
	if load != nil {
		if protocol, err := load(model); err == nil && validProtocol(protocol) {
			return protocol
		}
	}
	return initialProtocol(cfg)
}

func (c *Client) rememberProtocol(model, protocol string) {
	p := c.protocolState()
	p.mu.Lock()
	p.selected[model] = protocol
	save := p.save
	p.mu.Unlock()
	if save != nil {
		_ = save(model, protocol)
	}
}

func protocolForPath(path string) string {
	if path == "/responses" {
		return KindResponses
	}
	if path == "/chat/completions" {
		return KindOpenAI
	}
	return ""
}

func (c *Client) ActualProtocol(model string) string {
	cfg := c.Config()
	if cfg.Kind == KindOpenAI || cfg.Kind == KindResponses {
		return cfg.Kind
	}
	if cfg.Kind != KindAuto {
		return cfg.Kind
	}
	return c.protocolFor(model, cfg)
}
