package provider

type anthropicSystemBlock struct {
	Type         string          `json:"type"`
	Text         string          `json:"text"`
	CacheControl *anthropicCache `json:"cache_control,omitempty"`
}

// anthropicSystem sends a prompt with a CacheBreak as two text blocks, with
// the breakpoint at the end of the stable one. Without a marker it stays a
// plain string. With cache off no block carries cache_control.
func anthropicSystem(system string, cache bool) any {
	if !hasCacheBreak(system) {
		if system == "" {
			return nil
		}
		return system
	}
	stable, tail := SplitSystem(system)
	var blocks []anthropicSystemBlock
	if stable != "" {
		block := anthropicSystemBlock{Type: "text", Text: stable}
		if cache {
			block.CacheControl = &anthropicCache{Type: "ephemeral"}
		}
		blocks = append(blocks, block)
	}
	if tail != "" {
		blocks = append(blocks, anthropicSystemBlock{Type: "text", Text: tail})
	}
	if len(blocks) == 0 {
		return nil
	}
	return blocks
}
