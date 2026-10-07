package web

import "jin/internal/hooks"

// hookRef names a global hook, or with Project set a hook in .jin/hooks of
// the project Dir.
type hookRef struct {
	Name    string `json:"name"`
	Project bool   `json:"project"`
	Dir     string `json:"dir"`
	Content string `json:"content"`
	Trusted bool   `json:"trusted"`
}

func (h hookRef) path() (string, error) {
	if h.Project {
		return hooks.ProjectPath(h.Dir, h.Name)
	}
	return hooks.Path(h.Name)
}

func (h hookRef) key() string {
	if h.Project {
		return hooks.ProjectKey(h.Dir, h.Name)
	}
	return h.Name
}

type hookView struct {
	hookRef
	Enabled bool   `json:"enabled"`
	Preview string `json:"preview"`
}
