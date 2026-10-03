package provider

import (
	"encoding/json"
	"strings"
	"testing"
)

const testToolsSchema = `[{"type":"function","function":{"name":"read","description":"Read a file","parameters":{"type":"object","properties":{"path":{"type":"string"}}}}}]`

func TestResponsesPayload(t *testing.T) {
	history := []Message{
		{Role: "system", Content: "rules"},
		{Role: "user", Content: "look", Images: []Image{{MimeType: "image/png", Data: "AAAA"}}},
		{Role: "assistant", Content: "reading", ToolCalls: []ToolCall{testCall("call_1", "read", `{"path":"a"}`)}},
		{Role: "tool", ToolCallID: "call_1", Content: "file body"},
	}
	data, err := buildResponsesPayload("m", "high", history, json.RawMessage(testToolsSchema))
	if err != nil {
		t.Fatal(err)
	}
	var body struct {
		Store     *bool             `json:"store"`
		Stream    bool              `json:"stream"`
		Include   []string          `json:"include"`
		Reasoning map[string]string `json:"reasoning"`
		Input     []map[string]any  `json:"input"`
		Tools     []map[string]any  `json:"tools"`
	}
	if err := json.Unmarshal(data, &body); err != nil {
		t.Fatal(err)
	}
	if body.Store == nil || *body.Store || !body.Stream || strings.Join(body.Include, ",") != "reasoning.encrypted_content" {
		t.Fatalf("flags: %s", data)
	}
	if body.Reasoning["effort"] != "high" || body.Reasoning["summary"] != "auto" {
		t.Fatalf("reasoning: %v", body.Reasoning)
	}
	if len(body.Tools) != 1 || body.Tools[0]["name"] != "read" || body.Tools[0]["strict"] != false {
		t.Fatalf("tools: %v", body.Tools)
	}
	want := []string{"message/developer", "message/user", "message/assistant", "function_call/", "function_call_output/"}
	if len(body.Input) != len(want) {
		t.Fatalf("input: %s", data)
	}
	for i, item := range body.Input {
		role, _ := item["role"].(string)
		if got := item["type"].(string) + "/" + role; got != want[i] {
			t.Fatalf("item %d = %s, want %s", i, got, want[i])
		}
	}
	for _, needle := range []string{`"input_image"`, `"output_text"`, `"call_id":"call_1"`, `"output":"file body"`} {
		if !strings.Contains(string(data), needle) {
			t.Fatalf("missing %s in %s", needle, data)
		}
	}
	if strings.Contains(string(data), "jin_native") || strings.Contains(string(data), "previous_response_id") {
		t.Fatalf("unexpected field: %s", data)
	}
}

func TestResponsesPayloadNoEffort(t *testing.T) {
	data, err := buildResponsesPayload("m", "", []Message{{Role: "user", Content: "hi"}}, nil)
	if err != nil || strings.Contains(string(data), "reasoning\":") || strings.Contains(string(data), "tools") {
		t.Fatalf("payload: %s %v", data, err)
	}
	if _, err := responsesTools(json.RawMessage(`[{"type":"web"}]`)); err == nil {
		t.Fatal("non-function tool must fail")
	}
}

func testCall(id, name, args string) ToolCall {
	call := ToolCall{ID: id, Type: "function"}
	call.Function.Name, call.Function.Arguments = name, args
	return call
}
