package ui

import "jin/internal/provider"

func kindLabel(kind string) string {
	if kind == provider.KindAnthropic {
		return "Anthropic"
	}
	return "OpenAI-compatible"
}
