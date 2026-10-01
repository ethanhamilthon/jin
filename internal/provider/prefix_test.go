package provider

import (
	"bytes"
	"encoding/json"
	"testing"
)

func history() []Message {
	call := ToolCall{ID: "c1", Type: "function"}
	call.Function.Name, call.Function.Arguments = "read", `{"path": "a.go"}`
	return []Message{
		{Role: "system", Content: "sys"},
		{Role: "user", Content: "hi"},
		{Role: "assistant", ReasoningContent: "thinking", ToolCalls: []ToolCall{call}},
		{Role: "tool", ToolCallID: "c1", Content: "file"},
	}
}

func messagesJSON(t *testing.T, messages []Message) []byte {
	t.Helper()
	payload, err := chatPayload("m", "high", messages, json.RawMessage(`[{"a":1}]`))
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Messages json.RawMessage `json:"messages"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		t.Fatal(err)
	}
	return body.Messages
}

func TestRequestOnlyGrowsAtTheEnd(t *testing.T) {
	short := messagesJSON(t, history())
	long := messagesJSON(t, append(history(), Message{Role: "assistant", Content: "done"}))
	if !bytes.HasPrefix(long, bytes.TrimSuffix(short, []byte("]"))) {
		t.Fatalf("the earlier request is not a prefix of the later one:\n%s\n%s", short, long)
	}
}

func TestStoredMessagesSerializeIdentically(t *testing.T) {
	before := messagesJSON(t, history())
	var restored []Message
	for _, msg := range history() {
		data, _ := json.Marshal(msg)
		var back Message
		if err := json.Unmarshal(data, &back); err != nil {
			t.Fatal(err)
		}
		restored = append(restored, back)
	}
	if after := messagesJSON(t, restored); !bytes.Equal(before, after) {
		t.Fatalf("a stored message serializes differently after reload:\n%s\n%s", before, after)
	}
}
