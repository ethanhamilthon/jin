package ui

import (
	"strconv"
	"strings"

	"jin/internal/core"
)

type contextResult struct {
	label string
	bytes int
}

// contextInfo is what /context shows; sizes are in bytes.
type contextInfo struct {
	used, window int
	prompt       []core.PromptPart
	toolSchemas  int
	messages     int
	conversation int
	results      []contextResult
	cache        cacheRate
}

func approx(bytes int) string { return "~" + formatCount(core.EstimateTokens(bytes)) }

func (c contextInfo) String() string {
	var b strings.Builder
	b.WriteString("Context (token counts are approximate: 1 token ≈ 4 bytes)\n")
	b.WriteString("Used: " + formatCount(c.used))
	if c.window > 0 {
		b.WriteString(" of " + formatCount(c.window) + " (" + strconv.Itoa(c.used*100/c.window) + "%)")
	}
	b.WriteString("\nSystem prompt: " + approx(promptBytes(c.prompt)) + "\n")
	for _, part := range c.prompt {
		if part.Name != "cache break" {
			b.WriteString("  " + part.Name + ": " + approx(part.Bytes()) + "\n")
		}
	}
	b.WriteString("Tool schemas: " + approx(c.toolSchemas) + "\n")
	b.WriteString("Conversation: " + strconv.Itoa(c.messages) + " messages, " + approx(c.conversation))
	if len(c.results) > 0 {
		b.WriteString("\nLargest tool results:")
	}
	for _, r := range c.results {
		b.WriteString("\n  " + r.label + ": " + approx(r.bytes))
	}
	if c.cache.known {
		b.WriteString("\nCache: " + strconv.Itoa(c.cache.percent) + "% of the last request")
	}
	return b.String()
}

func promptBytes(parts []core.PromptPart) int {
	n := 0
	for _, part := range parts {
		if part.Name != "cache break" {
			n += part.Bytes()
		}
	}
	return n
}
