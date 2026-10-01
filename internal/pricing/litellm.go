package pricing

import "encoding/json"

type liteLLMEntry struct {
	InputCostPerToken       float64 `json:"input_cost_per_token"`
	OutputCostPerToken      float64 `json:"output_cost_per_token"`
	CacheReadInputTokenCost float64 `json:"cache_read_input_token_cost"`
	MaxInputTokens          int     `json:"max_input_tokens"`
	SupportsVision          *bool   `json:"supports_vision"`
}

// parseLiteLLM decodes entry by entry: the catalogue holds documentation
// records such as "sample_spec" whose fields have other types, and one such
// record must not discard the whole table.
func parseLiteLLM(data []byte) (Table, error) {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	table := make(Table, len(raw))
	for name, item := range raw {
		var entry liteLLMEntry
		if json.Unmarshal(item, &entry) != nil {
			continue
		}
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
