package provider

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// promptCacheKey names a prompt prefix for OpenAI's cache routing: the first
// 16 hex chars of the sha256 of the stable part of the system prompt. It is
// empty when there is no system prompt.
func promptCacheKey(messages []Message) string {
	var texts []string
	for _, m := range messages {
		if m.Role == "system" && strings.TrimSpace(m.Content) != "" {
			texts = append(texts, m.Content)
		}
	}
	stable, _ := SplitSystem(strings.Join(texts, "\n\n"))
	if stable == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(stable))
	return hex.EncodeToString(sum[:])[:16]
}
