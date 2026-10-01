package core

import (
	_ "embed"
	"runtime"
	"strings"
	"time"
)

//go:embed system_prompt.md
var systemPrompt string

func SystemPrompt(dir string) string {
	return strings.NewReplacer(
		"{{dir}}", dir,
		"{{os}}", runtime.GOOS+"/"+runtime.GOARCH,
		"{{date}}", time.Now().Format("2006-01-02"),
		"{{cat AGENTS.md}}", renderContext(ContextFiles(dir)),
	).Replace(systemPrompt)
}
