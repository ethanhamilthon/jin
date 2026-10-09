package core

import (
	"runtime"
	"slices"
)

const (
	macOSNotice    = "macOS has BSD userland: never run sed -i. Change source files with the edit tool; for bulk replacements use perl -pi -e."
	todoFileNotice = "For work with three or more steps, keep a checklist in TODO.md in the working directory and tick items as you finish them."
)

// notices are the short rules that depend on the machine and on the tools
// the agent has. They sit in the stable part of the prompt, before the cache
// break, because they only change when the tool set does.
func notices(in PromptInput) []PromptPart {
	var parts []PromptPart
	if in.OS == "" {
		in.OS = runtime.GOOS
	}
	if in.OS == "darwin" {
		parts = append(parts, PromptPart{"macOS notice", macOSNotice})
	}
	if slices.Contains(in.ToolNames, "write") || slices.Contains(in.ToolNames, "edit") {
		parts = append(parts, PromptPart{"todo file notice", todoFileNotice})
	}
	return parts
}
