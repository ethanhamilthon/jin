package provider

import (
	"regexp"
	"strconv"
	"strings"
)

const defaultAnthropicMaxTokens = 32000

var (
	inclusiveTokensPattern = regexp.MustCompile(
		`(?i)(?:\d+\s*>\s*|<=|less than\s+or\s+equal\s+to|at most|(?:not|cannot)\s+exceed|between\s+\d+\s+and|(?:maximum|limit)(?:\s+allowed)?(?:\s+is|\s+of)?)\s*(\d+)`,
	)
	strictTokensPattern = regexp.MustCompile(
		`(?i)(?:less than|below|under|<)\s*(\d+)`,
	)
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
	strict := false
	match := inclusiveTokensPattern.FindStringSubmatch(rest)
	if match == nil {
		match = strictTokensPattern.FindStringSubmatch(rest)
		strict = true
	}
	if match == nil {
		return 0, false
	}
	limit, err := strconv.Atoi(match[1])
	if err != nil {
		return 0, false
	}
	if strict {
		limit--
	}
	if limit <= 0 || limit >= requested {
		return 0, false
	}
	return limit, true
}
