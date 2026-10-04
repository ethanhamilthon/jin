package core

import (
	"runtime"
	"strings"
	"time"
)

// sessionTail is the part of the system prompt that changes from session to
// session: the environment and the session id. A system text that has its own
// "Environment:" line keeps it, so the block is not added twice.
func sessionTail(in PromptInput, system string) string {
	var lines []string
	if !hasEnvironmentLine(system) {
		lines = append(lines,
			"Environment:",
			"- Working directory: "+in.Dir,
			"- OS: "+runtime.GOOS+" "+runtime.GOARCH,
			"- Date: "+time.Now().Format("2006-01-02"),
		)
	}
	if in.SessionID != "" {
		lines = append(lines, "Your session id: "+in.SessionID)
	}
	return strings.Join(lines, "\n")
}

func hasEnvironmentLine(text string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "Environment:") {
			return true
		}
	}
	return false
}
