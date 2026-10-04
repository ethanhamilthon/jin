package store

import "testing"

func TestVoiceRoundTrip(t *testing.T) {
	db := openTest(t)
	if cfg, err := db.LoadConfig(); err != nil || cfg.Voice.Ready() {
		t.Fatalf("fresh config voice = %+v, err = %v", cfg.Voice, err)
	}
	want := Voice{BaseURL: "https://api.groq.com/openai/v1", APIKey: "k", Model: "whisper-large-v3-turbo", Language: "ru"}
	if err := db.SaveVoice(want); err != nil {
		t.Fatal(err)
	}
	cfg, err := db.LoadConfig()
	if err != nil || cfg.Voice != want || !cfg.Voice.Ready() {
		t.Fatalf("voice = %+v, err = %v", cfg.Voice, err)
	}
}
