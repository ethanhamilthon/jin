package ui

import "jin/internal/provider"

func kindLabel(kind string) string {
	switch kind {
	case provider.KindAnthropic:
		return "Anthropic-compatible"
	case provider.KindResponses:
		return "OpenAI Responses"
	default:
		return "OpenAI Chat Completions"
	}
}
