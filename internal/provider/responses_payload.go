package provider

import "encoding/json"

type responsesReasoning struct {
	Effort  string `json:"effort,omitempty"`
	Summary string `json:"summary,omitempty"`
}

type responsesPayload struct {
	Model     string              `json:"model"`
	Input     []json.RawMessage   `json:"input"`
	Tools     []responsesTool     `json:"tools,omitempty"`
	Reasoning *responsesReasoning `json:"reasoning,omitempty"`
	Include   []string            `json:"include"`
	Store     bool                `json:"store"`
	Stream    bool                `json:"stream"`
	CacheKey  string              `json:"prompt_cache_key,omitempty"`
}

func buildResponsesPayload(model, effort string, messages []Message, toolsSchema json.RawMessage) ([]byte, error) {
	tools, err := responsesTools(toolsSchema)
	if err != nil {
		return nil, err
	}
	p := responsesPayload{
		Model: model, Input: responsesHistory(withoutCacheBreak(nativeHistory(messages, KindResponses, model))),
		Tools: tools, Include: []string{"reasoning.encrypted_content"}, Stream: true,
		CacheKey: promptCacheKey(messages),
	}
	if effort != "" {
		p.Reasoning = &responsesReasoning{Effort: effort, Summary: "auto"}
	}
	return json.Marshal(p)
}
