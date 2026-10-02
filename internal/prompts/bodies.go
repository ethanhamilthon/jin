package prompts

import (
	"os"
	"slices"
	"strings"
)

// Bodies reads the raw text of every enabled prompt: the built-in ones and
// the files in ~/.jin/prompts. A prompt with no text is left out. The text is
// not filled in; running its {{commands}} is the job of the caller.
func Bodies(disabled []string) map[string]string {
	infos, _ := ListInfo()
	bodies := make(map[string]string, len(infos))
	for _, info := range infos {
		if slices.Contains(disabled, info.Name) {
			continue
		}
		var body string
		if info.System {
			body, _ = systemBody(info.Name)
		} else if path, err := Path(info.Name); err == nil {
			if data, err := os.ReadFile(path); err == nil {
				body = strings.TrimSpace(string(data))
			}
		}
		if body != "" {
			bodies[info.Name] = body
		}
	}
	return bodies
}
