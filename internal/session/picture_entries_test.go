package session

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/png"
	"strings"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/tools"
)

func TestHistoryNamesPicturesAfterTheirReadCalls(t *testing.T) {
	var buf bytes.Buffer
	_ = png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, 3, 5)))
	picture := provider.Image{MimeType: "image/png", Data: base64.StdEncoding.EncodeToString(buf.Bytes())}
	call := func(id, path string) provider.ToolCall {
		c := provider.ToolCall{ID: id, Type: "function"}
		c.Function.Name, c.Function.Arguments = "read", `{"path":"`+path+`"}`
		return c
	}
	messages := []provider.Message{
		{Role: "assistant", ToolCalls: []provider.ToolCall{call("a", "/x/one.png"), call("b", "/x/two.png")}},
		{Role: "tool", ToolCallID: "a", Content: "Read image file [image/png]"},
		{Role: "tool", ToolCallID: "b", Content: "Read image file [image/png]"},
		{Role: "user", Content: "Attached image(s) from tool result:", Images: []provider.Image{picture, picture}},
	}
	var pictures []Entry
	for _, e := range History(messages, tools.NewRegistry(tools.NewRead())) {
		if e.Picture {
			pictures = append(pictures, e)
		}
	}
	if len(pictures) != 2 || !strings.Contains(pictures[0].Text, "one.png") || !strings.Contains(pictures[1].Text, "two.png ") || !strings.HasSuffix(pictures[1].Text, "3×5]") {
		t.Fatalf("pictures = %+v", pictures)
	}
	if core.ImageLabels(messages[3]) == nil {
		t.Fatal("test message is not an image message")
	}
}
