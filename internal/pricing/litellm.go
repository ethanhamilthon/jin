package pricing

import "encoding/json"

func parseLiteLLM(data []byte) (Table, error) {
	var raw map[string]struct {
		InputCostPerToken       float64 `json:"input_cost_per_token"`
		OutputCostPerToken      float64 `json:"output_cost_per_token"`
		CacheReadInputTokenCost float64 `json:"cache_read_input_token_cost"`
		MaxInputTokens          int     `json:"max_input_tokens"`
		SupportsVision          *bool   `json:"supports_vision"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	table := make(Table, len(raw))
	for name, entry := range raw {
		table[name] = Entry{
			InputCostPerToken:     entry.InputCostPerToken,
			OutputCostPerToken:    entry.OutputCostPerToken,
			CacheReadCostPerToken: entry.CacheReadInputTokenCost,
			MaxInputTokens:        entry.MaxInputTokens,
			VisionKnown:           entry.SupportsVision != nil,
			Vision:                entry.SupportsVision != nil && *entry.SupportsVision,
		}
	}
	return table, nil
}
