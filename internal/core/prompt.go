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

//go:embed jin_docs_prompt.md
var jinDocsPrompt string

// SystemPrompt fills the template. The jin docs pointer (when switched on)
// and the enabled hooks go in as plain text right before the AGENTS.md block.
func SystemPrompt(dir string, disabledHooks []string, jinDocs bool) string {
	docs := ""
	if jinDocs {
		docs = strings.TrimSpace(jinDocsPrompt)
	}
	return strings.NewReplacer(
		"{{jin_docs}}", optionalBlock(docs),
		"{{hooks}}", optionalBlock(hooks.Render(disabledHooks)),
		"{{dir}}", dir,
		"{{os}}", runtime.GOOS+"/"+runtime.GOARCH,
		"{{date}}", time.Now().Format("2006-01-02"),
		"{{cat AGENTS.md}}", renderContext(ContextFiles(dir)),
	).Replace(systemPrompt)
}

func optionalBlock(rendered string) string {
	if rendered == "" {
		return ""
	}
	return rendered + "\n\n"
}
