package provider

import "encoding/json"

type responsesEvent struct {
	Type     string          `json:"type"`
	Delta    string          `json:"delta"`
	Item     json.RawMessage `json:"item"`
	Response *responsesFinal `json:"response"`
	Code     string          `json:"code"`
	Message  string          `json:"message"`
}

type responsesFinal struct {
	ID                string            `json:"id"`
	Model             string            `json:"model"`
	Output            []json.RawMessage `json:"output"`
	Usage             *responsesUsage   `json:"usage"`
	Error             *responsesError   `json:"error"`
	IncompleteDetails *struct {
		Reason string `json:"reason"`
	} `json:"incomplete_details"`
}

type responsesError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type responsesUsage struct {
	Input        int `json:"input_tokens"`
	Output       int `json:"output_tokens"`
	InputDetails *struct {
		Cached     *int `json:"cached_tokens"`
		CacheWrite *int `json:"cache_write_tokens"`
	} `json:"input_tokens_details"`
	OutputDetails *struct {
		Reasoning *int `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
}

func (u *responsesUsage) usage() Usage {
	if u == nil {
		return Usage{}
	}
	out := Usage{Input: u.Input, Output: u.Output, Known: true}
	if u.InputDetails != nil {
		if n := u.InputDetails.Cached; n != nil {
			out.CachedInput, out.CacheKnown = *n, true
		}
		if n := u.InputDetails.CacheWrite; n != nil {
			out.CacheWriteInput, out.CacheWriteKnown = *n, true
		}
	}
	if u.OutputDetails != nil && u.OutputDetails.Reasoning != nil {
		out.Reasoning, out.ReasoningKnown = *u.OutputDetails.Reasoning, true
	}
	return out
}
