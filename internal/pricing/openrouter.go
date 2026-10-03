package pricing

import (
	"encoding/json"
	"slices"
	"strconv"
)

// parseOpenRouter reads https://openrouter.ai/api/v1/models. Prices arrive as
// decimal strings in dollars per token, already matching Entry's units.
func parseOpenRouter(data []byte) (Table, error) {
	var raw struct {
		Data []struct {
			ID            string   `json:"id"`
			ContextLength int      `json:"context_length"`
			Parameters    []string `json:"supported_parameters"`
			Architecture  struct {
				InputModalities []string `json:"input_modalities"`
			} `json:"architecture"`
			Pricing struct {
				Prompt         string `json:"prompt"`
				Completion     string `json:"completion"`
				InputCacheRead string `json:"input_cache_read"`
			} `json:"pricing"`
		} `json:"data"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	table := make(Table, len(raw.Data))
	for _, model := range raw.Data {
		table[model.ID] = Entry{
			InputCostPerToken:     parseDollars(model.Pricing.Prompt),
			OutputCostPerToken:    parseDollars(model.Pricing.Completion),
			CacheReadCostPerToken: parseDollars(model.Pricing.InputCacheRead),
			MaxInputTokens:        model.ContextLength,
			VisionKnown:           len(model.Architecture.InputModalities) > 0,
			Vision:                slices.Contains(model.Architecture.InputModalities, "image"),
			ReasoningKnown:        len(model.Parameters) > 0,
			Reasoning:             slices.Contains(model.Parameters, "reasoning"),
		}
	}
	return table, nil
}

func parseDollars(value string) float64 {
	amount, _ := strconv.ParseFloat(value, 64)
	return amount
}
