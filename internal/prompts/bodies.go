package prompts

import (
	"errors"
	"io/fs"
	"os"
	"slices"
	"strings"
)

// Bodies reads the raw text of every enabled prompt: the built-in ones and
// the files in ~/.jin/prompts. A prompt with no text is left out. The text is
// not filled in; running its {{commands}} is the job of the caller.
func Bodies(disabled []string) map[string]string {
	bodies, _ := LoadBodies(disabled)
	return bodies
}

// LoadBodies is Bodies plus a warning for each prompt file or folder that
// exists but cannot be read. The built-in prompts are always listed.
func LoadBodies(disabled []string) (map[string]string, []string) {
	infos, err := ListInfo()
	var warnings []string
	if err != nil {
		warnings = append(warnings, "cannot list prompts: "+err.Error())
		infos = slices.Clone(systemPromptList)
	}
	bodies := make(map[string]string, len(infos))
	for _, info := range infos {
		if slices.Contains(disabled, info.Name) {
			continue
		}
		var body string
		if info.System {
			body, _ = systemBody(info.Name)
		} else if path, err := Path(info.Name); err == nil {
			data, err := os.ReadFile(path)
			if err != nil && !errors.Is(err, fs.ErrNotExist) {
				warnings = append(warnings, "cannot read prompt: "+err.Error())
			}
			body = strings.TrimSpace(string(data))
		}
		if body != "" {
			bodies[info.Name] = body
		}
	}
	return bodies, warnings
}
