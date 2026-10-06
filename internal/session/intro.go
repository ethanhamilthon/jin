package session

import (
	"slices"

	"jin/internal/core"
	"jin/internal/hooks"
	"jin/internal/prompts"
	"jin/internal/store"
	"jin/internal/sysprompt"
	"jin/internal/tools"
)

// Intro is what a new session shows before its first message.
type Intro struct {
	Version      string   `json:"version"`
	Tools        []string `json:"tools"`
	Context      []string `json:"context"`
	Hooks        []string `json:"hooks"`
	Prompts      []string `json:"prompts"`
	SystemPrompt string   `json:"system_prompt,omitempty"`
	Update       string   `json:"update,omitempty"`
}

func (m *Manager) intro(cfg store.Config, dir string, _ []string) *Intro {
	in := &Intro{Version: m.version, Tools: tools.Without(cfg.ToolsDisabled), Update: m.latest}
	for _, f := range core.ContextFiles(dir) {
		in.Context = append(in.Context, f.Path+" ("+string(f.Kind)+")")
	}
	trust, _ := m.db.HooksTrust(dir)
	for _, hook := range hooks.ActiveIn(dir, cfg.HooksDisabled, trust == store.Trusted) {
		name := hook.Name
		if hook.Project {
			name += " (project)"
		}
		in.Hooks = append(in.Hooks, name)
	}
	infos, _ := prompts.ListInfo()
	for _, info := range infos {
		if !slices.Contains(cfg.PromptsDisabled, info.Name) {
			in.Prompts = append(in.Prompts, info.Name)
		}
	}
	if loaded, _ := sysprompt.Load(); loaded.Custom {
		in.SystemPrompt, _ = sysprompt.Path()
	}
	return in
}

// SetLatest records a newer release; new sessions mention it.
func (m *Manager) SetLatest(tag string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.latest = tag
}
