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
	tools        []core.PromptPart
	full         bool
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
	if c.full {
		c.writeTexts(&b)
	} else {
		b.WriteString("\n/context full prints the text of every part")
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

// writeTexts adds the exact text of each part of the system prompt and each
// tool schema, as it goes to the model.
func (c contextInfo) writeTexts(b *strings.Builder) {
	for _, part := range c.prompt {
		if part.Name != "cache break" {
			b.WriteString("\n\n--- " + part.Name + " ---\n" + part.Text)
		}
	}
	for _, tool := range c.tools {
		b.WriteString("\n\n--- tool schema: " + tool.Name + " ---\n" + tool.Text)
	}
}
