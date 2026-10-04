package ui

import (
	"os"
	"path/filepath"
	"testing"

	"jin/internal/core"
	"jin/internal/store"
)

func demoApp(t *testing.T) *app {
	a, _ := layoutApp(t)
	db, _ := openFoldDB(t)
	home, err := filepath.EvalSymlinks(os.Getenv("HOME"))
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", home)
	a.store, a.width = db, 110
	a.screen.SetSize(110, 32)
	a.cfg.Model, a.cfg.Effort, a.cfg.Motion = "gpt-6-luna", "xhigh", "normal"
	a.cfg.Providers = []store.ProviderEntry{{ID: "sample", Name: "Sample provider", BaseURL: "https://example.invalid/v1", APIKey: "demo-only"}}
	a.cfg.ActiveProvider = "sample"
	a.asyncRunning = map[string]int{}
	fixtures := []struct{ id, project, title, message string }{
		{"jin-chat", "jin", "Workspace redesign", "## Jin v0.8.0\nProjects keep one directory per session.\n\nThe shared input follows the focused pane."},
		{"site-chat", "website", "Build the blog", "## Website\nThe layout is ready. Next: write the first article."},
		{"research-chat", "research", "Review the notes", "## Research\nNotes and references stay in this project's session."},
		{"jin-review", "jin", "Release checks", "## Release checks\nTests, project isolation, and session ownership are covered."},
	}
	for _, fixture := range fixtures {
		dir := filepath.Join(os.Getenv("HOME"), "projects", fixture.project)
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		s := &chatSession{id: fixture.id, path: dir, title: fixture.title, model: a.cfg.Model,
			effort: a.cfg.Effort, store: db, provider: "sample", ready: true, persisted: true, width: 110,
			history: []chatEntry{{kind: core.UpdateAssistant, text: fixture.message}}}
		if err := db.TouchProvider(s.id, dir, s.model, s.effort, s.title, s.provider); err != nil {
			t.Fatal(err)
		}
		a.sessions[s.id] = s
	}
	a.active = a.sessions["jin-chat"]
	a.dir = a.active.path
	a.initPanes()
	return a
}

func installDemoPanes(a *app) {
	one := &paneNode{session: a.sessions["jin-chat"]}
	two := &paneNode{session: a.sessions["site-chat"]}
	three := &paneNode{session: a.sessions["research-chat"]}
	four := &paneNode{session: a.sessions["jin-review"]}
	a.panes = &paneNode{vertical: true,
		first:  &paneNode{first: one, second: three},
		second: &paneNode{first: two, second: four}}
	a.focused = one
}
