package ui

import (
	"strconv"
	"strings"

	"jin/internal/pricing"
)

// modelOptions are the rows of /model: the id, and after it what the
// catalogues know of the model: context window, price in and out per 1M
// tokens, and whether it reasons and sees images.
func modelOptions(models []string, table pricing.Table) []option {
	options := make([]option, len(models))
	for i, model := range models {
		options[i] = option{label: model, value: model}
		if entry, ok := table.Lookup(model); ok {
			options[i].detail = modelFacts(entry)
		}
	}
	return options
}

func modelFacts(e pricing.Entry) string {
	var facts []string
	if e.MaxInputTokens > 0 {
		facts = append(facts, formatCount(e.MaxInputTokens)+" ctx")
	}
	if e.InputCostPerToken > 0 || e.OutputCostPerToken > 0 {
		facts = append(facts, "$"+perMillion(e.InputCostPerToken)+" / $"+perMillion(e.OutputCostPerToken))
	}
	if e.ReasoningKnown && e.Reasoning {
		facts = append(facts, "reasoning")
	}
	if e.VisionKnown && e.Vision {
		facts = append(facts, "vision")
	}
	return strings.Join(facts, " · ")
}

func perMillion(perToken float64) string {
	value := perToken * 1_000_000
	if value >= 10 {
		return strconv.FormatFloat(value, 'f', 0, 64)
	}
	return strings.TrimRight(strings.TrimRight(strconv.FormatFloat(value, 'f', 2, 64), "0"), ".")
}
