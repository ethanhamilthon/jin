package provider

import (
	"encoding/json"
	"errors"
	"sort"
)

type anthropicBlocks map[int]map[string]json.RawMessage

func (blocks anthropicBlocks) observe(ev anthropicEvent, raw []byte) {
	if ev.Type == "content_block_start" {
		var start struct {
			Block map[string]json.RawMessage `json:"content_block"`
		}
		if json.Unmarshal(raw, &start) == nil && start.Block != nil {
			blocks[ev.Index] = start.Block
		}
	}
	if ev.Type != "content_block_delta" || ev.Delta == nil {
		return
	}
	block := blocks[ev.Index]
	if block == nil {
		return
	}
	var field, delta string
	switch ev.Delta.Type {
	case "text_delta":
		field, delta = "text", ev.Delta.Text
	case "thinking_delta":
		field, delta = "thinking", ev.Delta.Thinking
	case "signature_delta":
		field, delta = "signature", ev.Delta.Signature
	default:
		return
	}
	var current string
	_ = json.Unmarshal(block[field], &current)
	block[field], _ = json.Marshal(current + delta)
}

func (blocks anthropicBlocks) finish(msg *Message, calls map[int]*ToolCall) error {
	if len(blocks) == 0 {
		return nil
	}
	indices := make([]int, 0, len(blocks))
	for index := range blocks {
		indices = append(indices, index)
	}
	sort.Ints(indices)
	native := &NativeMessage{Kind: KindAnthropic}
	for _, index := range indices {
		block := blocks[index]
		if call := calls[index]; call != nil {
			if call.Function.Arguments == "" {
				call.Function.Arguments = "{}"
			}
			if !json.Valid([]byte(call.Function.Arguments)) {
				return errors.New("invalid Anthropic tool arguments")
			}
			block["input"] = json.RawMessage(call.Function.Arguments)
		}
		raw, err := json.Marshal(block)
		if err != nil {
			return err
		}
		native.Items = append(native.Items, raw)
	}
	msg.Native = native
	return nil
}
