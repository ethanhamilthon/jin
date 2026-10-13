package startup

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"jin/internal/core"
	"jin/internal/docs"
	"jin/internal/hooks"
	"jin/internal/provider"
)

// systemParts builds the system prompt in code: the built-in text, the pointer to
// the documentation, the hooks, AGENTS.md, a cache break and the environment.
// No shell command runs here; only the commands inside hooks run.
func systemParts(ctx context.Context, in Input, base string, env []string) ([]core.PromptPart, []string) {
	parts := []core.PromptPart{{Name: "system text", Text: base + "\n\n"}, {Name: "docs", Text: docs.Pointer + "\n\n"}}
	var warnings []string
	if in.Hooks != nil {
		text, hookWarnings := hooks.Fill(ctx, in.Dir, in.Hooks(), env)
		warnings = hookWarnings
		if text != "" {
			parts = append(parts, core.PromptPart{Name: "hooks", Text: text + "\n\n"})
		}
	}
	if text := agentsText(in.Dir); text != "" {
		parts = append(parts, core.PromptPart{Name: "AGENTS.md", Text: text + "\n\n"})
	}
	parts = append(parts,
		core.PromptPart{Name: "cache break", Text: provider.CacheBreak + "\n\n"},
		core.PromptPart{Name: "environment", Text: environment(in)},
	)
	return parts, warnings
}

func agentsText(dir string) string {
	data, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil || len(data) == 0 {
		return ""
	}
	return "AGENTS.md:\n\n" + strings.TrimRight(string(data), "\n")
}

func environment(in Input) string {
	lines := []string{
		"Environment:",
		"- Working directory: " + in.Dir,
		"- OS: " + osName(),
		"- Date: " + time.Now().Format("2006-01-02"),
	}
	if in.SessionID != "" {
		lines = append(lines, "Your session id: "+in.SessionID)
	}
	return strings.Join(lines, "\n")
}

func osName() string {
	system, arch := map[string]string{"darwin": "Darwin", "linux": "Linux"}[runtime.GOOS], runtime.GOARCH
	if system == "" {
		system = runtime.GOOS
	}
	switch {
	case arch == "amd64":
		arch = "x86_64"
	case arch == "arm64" && runtime.GOOS == "linux":
		arch = "aarch64"
	}
	return system + " " + arch
}
