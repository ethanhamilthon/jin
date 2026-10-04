package ui

import (
	"testing"

	"jin/internal/pricing"
)

func TestWindowFallsBackToSetting(t *testing.T) {
	db, _ := openFoldDB(t)
	s := &chatSession{store: db, model: "local/llama", pricing: pricing.Table{"known": {MaxInputTokens: 200_000}}}
	if got := s.window(); got != 0 {
		t.Fatalf("unknown window = %d, want 0", got)
	}
	cases := []struct {
		value string
		want  int
	}{{"32768", 32768}, {" 8000 ", 8000}, {"lots", 0}, {"-5", 0}}
	for _, c := range cases {
		if err := db.SetSetting("models.window.local/llama", c.value); err != nil {
			t.Fatal(err)
		}
		if got := s.window(); got != c.want {
			t.Errorf("setting %q: window = %d, want %d", c.value, got, c.want)
		}
	}
	_ = db.SetSetting("models.window.known", "1000")
	s.model = "known"
	if got := s.window(); got != 200_000 {
		t.Errorf("catalogue window = %d, want 200000", got)
	}
}
