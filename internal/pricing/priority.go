package pricing

import "strings"

func candidatePriority(name string) int {
	provider := name
	if idx := strings.Index(name, "/"); idx >= 0 {
		provider = name[:idx]
	}
	switch provider {
	case "openai":
		return 0
	case "anthropic":
		return 1
	case "google":
		return 2
	default:
		return 3
	}
}
