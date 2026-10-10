package store

import (
	"testing"

	"jin/internal/provider"
)

func TestTitleSettingsRoundTrip(t *testing.T) {
	db, _ := openTwo(t)
	cfg, _ := db.LoadConfig()
	if cfg.Title.After != DefaultTitleAfter || cfg.Title.Prompt != DefaultTitlePrompt || cfg.Title.Model != "" {
		t.Fatalf("defaults = %+v", cfg.Title)
	}
	if err := db.SaveTitle(TitleSettings{Provider: "p", Model: "m", Effort: "low", Prompt: "Name it.", After: 0}); err != nil {
		t.Fatal(err)
	}
	if cfg, _ := db.LoadConfig(); cfg.Title != (TitleSettings{Provider: "p", Model: "m", Effort: "low", Prompt: "Name it.", After: 0}) {
		t.Fatalf("saved = %+v", cfg.Title)
	}
	if cfg.Title.Refresh {
		t.Fatal("refresh is on by default")
	}
	if err := db.SaveTitle(TitleSettings{After: 2, Refresh: true, Prompt: DefaultTitlePrompt}); err != nil {
		t.Fatal(err)
	}
	if cfg, _ := db.LoadConfig(); cfg.Title.After != 2 || !cfg.Title.Refresh || cfg.Title.Prompt != DefaultTitlePrompt {
		t.Fatalf("default prompt = %+v", cfg.Title)
	}
}

func TestSaveTitleRejectsBadValues(t *testing.T) {
	db, _ := openTwo(t)
	if err := db.SaveTitle(TitleSettings{Model: "m", After: 4}); err == nil {
		t.Error("a model without a provider was saved")
	}
	for _, after := range []int{-1, maxTitleAfter + 1} {
		if err := db.SaveTitle(TitleSettings{After: after}); err == nil {
			t.Errorf("after %d was saved", after)
		}
	}
}

func TestSetTitle(t *testing.T) {
	db, _ := openTwo(t)
	if err := db.TouchProvider("s1", "", "m", "", "first prompt", ""); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"a", "b"} {
		if err := db.AppendMessage("s1", provider.Message{Role: "user", Content: text}); err != nil {
			t.Fatal(err)
		}
	}
	if err := db.SetTitle("s1", "Renamed"); err != nil {
		t.Fatal(err)
	}
	rec, ok, err := db.GetSession("s1")
	if err != nil || !ok || rec.Title != "Renamed" {
		t.Fatalf("session = %+v %v %v", rec, ok, err)
	}
}
