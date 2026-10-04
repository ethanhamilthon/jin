package voice

import (
	"context"
	"encoding/json"
	"testing"
)

func TestFalLanguageAndTask(t *testing.T) {
	for _, lang := range []string{"", "ru", "kk"} {
		c := Client{BaseURL: "https://fal.run", Model: "wizper", Language: lang}
		req, err := c.falRequest(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}
		var body map[string]any
		if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["task"] != "transcribe" {
			t.Fatalf("task=%v", body["task"])
		}
		got, exists := body["language"]
		if !exists || (lang == "" && got != nil) || (lang != "" && got != lang) {
			t.Fatalf("language=%q got=%v exists=%v", lang, got, exists)
		}
	}
}
