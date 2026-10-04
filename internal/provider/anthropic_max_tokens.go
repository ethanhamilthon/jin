package provider

import (
	"regexp"
	"strconv"
)

const defaultAnthropicMaxTokens = 32000

// Matches "max_tokens: 40000 > 8192" and "max_tokens must be <= 8192" style
// rejections, but not context overflows such as "input + max_tokens > window".
var maxTokensPattern = regexp.MustCompile(`max_tokens\D*?(?:\d+\s*>\s*|(?:<=|less than or equal to|at most)\s*)(\d+)`)

func maxTokensLimit(text string, requested int) (int, bool) {
	match := maxTokensPattern.FindStringSubmatch(text)
	if match == nil {
		return 0, false
	}
	limit, err := strconv.Atoi(match[1])
	if err != nil || limit <= 0 || limit >= requested {
		return 0, false
	}
	return limit, true
}
