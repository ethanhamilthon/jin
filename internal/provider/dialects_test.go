package provider

import (
	"encoding/json"
	"strings"
	"testing"
)

func nativeTurn(kind, model, item string) []Message {
	return []Message{
		{Role: "user", Content: "hi"},
		{Role: "assistant", Content: "fallback", Native: &NativeMessage{Kind: kind, Model: model,
			Items: []json.RawMessage{json.RawMessage(item)}}},
	}
}

func TestNativeRoundTripJSON(t *testing.T) {
	in := nativeTurn(KindResponses, "m", `{"type":"reasoning","encrypted_content":"x"}`)[1]
	data, err := json.Marshal(in)
	if err != nil || !strings.Contains(string(data), `"jin_native"`) {
		t.Fatalf("marshal: %s %v", data, err)
	}
	var out Message
	if err := json.Unmarshal(data, &out); err != nil || out.Native == nil || string(out.Native.Items[0]) != string(in.Native.Items[0]) {
		t.Fatalf("roundtrip: %+v %v", out, err)
	}
}

func TestResponsesReplaysNativeSameModel(t *testing.T) {
	item := `{"type":"reasoning","id":"rs_1","encrypted_content":"secret-blob","summary":[]}`
	data, _ := buildResponsesPayload("m", "", nativeTurn(KindResponses, "m", item), nil)
	if !strings.Contains(string(data), "secret-blob") || strings.Contains(string(data), "fallback") {
		t.Fatalf("same model must replay native: %s", data)
	}
	data, _ = buildResponsesPayload("other", "", nativeTurn(KindResponses, "m", item), nil)
	if strings.Contains(string(data), "secret-blob") || !strings.Contains(string(data), "fallback") {
		t.Fatalf("other model must use generic history: %s", data)
	}
}

func TestAnthropicNativeFiltering(t *testing.T) {
	item := `{"type":"thinking","thinking":"t","signature":"sig-1"}`
	data, _ := buildAnthropicPayload("c", "", nativeTurn(KindAnthropic, "c", item), nil, anthropicOptions{thinking: true, cache: true})
	if !strings.Contains(string(data), "sig-1") {
		t.Fatalf("same model must replay signature: %s", data)
	}
	data, _ = buildAnthropicPayload("c", "", nativeTurn(KindResponses, "c", item), nil, anthropicOptions{thinking: true, cache: true})
	if strings.Contains(string(data), "sig-1") || !strings.Contains(string(data), "fallback") {
		t.Fatalf("foreign native must not leak: %s", data)
	}
	if !strings.Contains(string(data), `"cache_control":{"type":"ephemeral"}`) {
		t.Fatalf("missing cache_control: %s", data)
	}
}

func TestChatStripsNative(t *testing.T) {
	history := nativeTurn(KindResponses, "m", `{"type":"reasoning"}`)
	data, err := chatPayload("m", "", history, nil)
	if err != nil || strings.Contains(string(data), "jin_native") || !strings.Contains(string(data), "fallback") {
		t.Fatalf("chat payload: %s %v", data, err)
	}
	if history[1].Native == nil {
		t.Fatal("chatPayload must not mutate history")
	}
}
