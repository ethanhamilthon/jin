package provider

func protocolForPath(path string) string {
	switch path {
	case "/responses":
		return KindResponses
	case "/chat/completions":
		return KindOpenAI
	}
	return ""
}

func (c *Client) ActualProtocol(model string) string {
	cfg := c.Config()
	if cfg.Kind != KindAuto {
		return cfg.Kind
	}
	return c.protocolFor(model, cfg)
}
