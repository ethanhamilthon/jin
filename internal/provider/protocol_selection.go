package provider

import "sync"

type protocolKey struct {
	kind, protocol, baseURL, model string
}

func keyFor(cfg Config, model string) protocolKey {
	return protocolKey{kind: cfg.Kind, protocol: cfg.Protocol, baseURL: NormalizeBaseURL(cfg.BaseURL), model: model}
}

type protocolState struct {
	mu       sync.RWMutex
	selected map[protocolKey]string
	load     func(string) (string, error)
	save     func(string, string) error
}

func (c *Client) protocolState() *protocolState {
	c.protocolMu.Lock()
	defer c.protocolMu.Unlock()
	if c.protocols == nil {
		c.protocols = &protocolState{selected: map[protocolKey]string{}}
	}
	return c.protocols
}

func (c *Client) SetProtocolStore(load func(model string) (string, error), save func(model, protocol string) error) {
	p := c.protocolState()
	p.mu.Lock()
	p.load, p.save = load, save
	p.selected = map[protocolKey]string{}
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
	key := keyFor(cfg, model)
	p.mu.RLock()
	load, local := p.load, p.selected[key]
	p.mu.RUnlock()
	if load == nil {
		if local != "" {
			return local
		}
		return initialProtocol(cfg)
	}
	protocol, err := load(model)
	if err != nil {
		if local != "" {
			return local
		}
	} else {
		p.forget(key)
		if validProtocol(protocol) {
			return protocol
		}
	}
	return initialProtocol(cfg)
}

func (p *protocolState) forget(key protocolKey) {
	p.mu.Lock()
	delete(p.selected, key)
	p.mu.Unlock()
}

func (c *Client) rememberProtocol(cfg Config, model, protocol string) {
	if keyFor(c.Config(), model) != keyFor(cfg, model) {
		return
	}
	p := c.protocolState()
	key := keyFor(cfg, model)
	p.mu.Lock()
	p.selected[key] = protocol
	save := p.save
	p.mu.Unlock()
	if save != nil && save(model, protocol) == nil && p.load != nil {
		p.forget(key)
	}
}
