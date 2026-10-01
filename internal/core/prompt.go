package core

import (
	_ "embed"
	"runtime"
	"strings"
	"time"

	"jin/internal/hooks"
)

//go:embed system_prompt.md
var systemPrompt string

// SystemPrompt fills the template. The enabled hooks, all but the disabled
// ones, go in as plain text right before the AGENTS.md block.
func SystemPrompt(dir string, disabledHooks []string) string {
	return strings.NewReplacer(
		"{{hooks}}", hooksBlock(hooks.Render(disabledHooks)),
		"{{dir}}", dir,
		"{{os}}", runtime.GOOS+"/"+runtime.GOARCH,
		"{{date}}", time.Now().Format("2006-01-02"),
		"{{cat AGENTS.md}}", renderContext(ContextFiles(dir)),
	).Replace(systemPrompt)
}

func hooksBlock(rendered string) string {
	if rendered == "" {
		return ""
	}
	return rendered + "\n\n"
}
