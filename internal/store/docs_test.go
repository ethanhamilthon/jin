package store

import "testing"

func TestJinDocsRoundTrip(t *testing.T) {
	db, _ := openTwo(t)
	if cfg, _ := db.LoadConfig(); !cfg.JinDocs {
		t.Fatal("jin docs should start on")
	}
	if err := db.SaveJinDocs(false); err != nil {
		t.Fatal(err)
	}
	if cfg, _ := db.LoadConfig(); cfg.JinDocs {
		t.Error("saved off, loaded on")
	}
	if err := db.SaveJinDocs(true); err != nil {
		t.Fatal(err)
	}
	if cfg, _ := db.LoadConfig(); !cfg.JinDocs {
		t.Error("saved on, loaded off")
	}
}
