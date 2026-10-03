package store

import (
	"encoding/json"
	"testing"

	"jin/internal/provider"
)

func TestNativeMessagePersists(t *testing.T) {
	db := openTest(t)
	item := `{"type":"reasoning","id":"rs_1","encrypted_content":"blob","summary":[]}`
	msg := provider.Message{Role: "assistant", Content: "hi", Native: &provider.NativeMessage{
		Kind: provider.KindResponses, Model: "m", Items: []json.RawMessage{json.RawMessage(item)},
	}}
	if err := db.AppendMessage("s1", msg); err != nil {
		t.Fatal(err)
	}
	got, err := db.LoadMessages("s1")
	if err != nil || len(got) != 1 || got[0].Native == nil {
		t.Fatalf("load: %+v %v", got, err)
	}
	n := got[0].Native
	if n.Kind != provider.KindResponses || n.Model != "m" || len(n.Items) != 1 || string(n.Items[0]) != item {
		t.Fatalf("native = %+v", n)
	}
}
