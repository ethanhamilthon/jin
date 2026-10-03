package provider

import (
	"strings"
	"time"
)

func effortTimeout(effort string) time.Duration {
	switch strings.ToLower(strings.TrimSpace(effort)) {
	case "low":
		return 120 * time.Second
	case "medium":
		return 180 * time.Second
	case "high":
		return 300 * time.Second
	case "xhigh", "max":
		return 600 * time.Second
	default:
		return DefaultStallTimeout
	}
}
