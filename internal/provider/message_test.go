package provider

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func TestImageMessageUsesContentParts(t *testing.T) {
	msg := Message{Role: "user", Content: "look", Images: []Image{{MimeType: "image/png", Data: "QUJD"}}}
	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"role":"user","content":[{"type":"text","text":"look"},{"type":"image_url","image_url":{"url":"data:image/png;base64,QUJD"}}]}`
	if string(data) != want {
		t.Fatalf("got  %s\nwant %s", data, want)
	}
	var back Message
	if err := json.Unmarshal(data, &back); err != nil {
		t.Fatal(err)
	}
	if back.Content != "look" || len(back.Images) != 1 || back.Images[0] != msg.Images[0] {
		t.Fatalf("round trip = %+v", back)
	}
	again, _ := json.Marshal(back)
	if !bytes.Equal(data, again) {
		t.Fatalf("serializes differently after reload:\n%s\n%s", data, again)
	}
}

func TestPlainMessageKeepsStringContent(t *testing.T) {
	data, _ := json.Marshal(Message{Role: "user", Content: "hi"})
	if string(data) != `{"role":"user","content":"hi"}` {
		t.Fatalf("got %s", data)
	}
	var back Message
	if err := json.Unmarshal(data, &back); err != nil || back.Content != "hi" || back.Images != nil {
		t.Fatalf("back = %+v, err = %v", back, err)
	}
}

func TestImageLabel(t *testing.T) {
	label := Image{MimeType: "image/png", Data: strings.Repeat("A", 4096)}.Label()
	if label != "image/png · 3 KB" {
		t.Fatalf("label = %q", label)
	}
}
