package provider

import (
	"slices"
	"strings"
)

// CacheBreak is the line that ends the stable part of a system prompt. What
// comes after it (the environment and the session id) changes from session
// to session. Adapters cut the prompt here and never send the line itself.
const CacheBreak = "<<jin-cache-break>>"

// SplitSystem cuts a system prompt at its last CacheBreak. Without the
// marker the whole text is stable.
func SplitSystem(system string) (stable, tail string) {
	if at := strings.LastIndex(system, CacheBreak); at >= 0 {
		stable, tail = system[:at], system[at+len(CacheBreak):]
	} else {
		stable = system
	}
	stable = strings.ReplaceAll(stable, CacheBreak, "")
	return strings.TrimSpace(stable), strings.TrimSpace(tail)
}

func hasCacheBreak(system string) bool { return strings.Contains(system, CacheBreak) }

// joinSystem is the prompt as one text, for APIs that take no blocks.
func joinSystem(system string) string {
	stable, tail := SplitSystem(system)
	return strings.TrimSpace(stable + "\n\n" + tail)
}

// withoutCacheBreak returns messages whose system texts carry no marker.
func withoutCacheBreak(messages []Message) []Message {
	out := slices.Clone(messages)
	for i := range out {
		if out[i].Role == "system" && hasCacheBreak(out[i].Content) {
			out[i].Content = joinSystem(out[i].Content)
		}
	}
	return out
}
