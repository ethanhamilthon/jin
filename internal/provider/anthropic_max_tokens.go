package provider

import (
	"regexp"
	"strconv"
	"strings"
)

const defaultAnthropicMaxTokens = 32000

var maxTokensPattern = regexp.MustCompile(
	`(?i)(?:\d+\s*>\s*|<=|less than(?: or equal to)?|at most|(?:not|cannot)\s+exceed|between\s+\d+\s+and|(?:maximum|limit)(?:\s+allowed)?(?:\s+is|\s+of)?)\s*(\d+)`,
)

func maxTokensLimit(text string, requested int) (int, bool) {
	lower := strings.ToLower(text)
	if strings.Contains(lower, "context") || strings.Contains(text, "+") {
		return 0, false
	}
	_, rest, found := strings.Cut(lower, "max_tokens")
	if !found {
		return 0, false
	}
	match := maxTokensPattern.FindStringSubmatch(rest)
	if match == nil {
		return 0, false
	}
	limit, err := strconv.Atoi(match[1])
	if err != nil || limit <= 0 || limit >= requested {
		return 0, false
	}
	return limit, true
}
