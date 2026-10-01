package ui

import "jin/internal/core"

// needsGap reports whether a blank row separates entry from the entry before
// it. Consecutive tool calls stay packed; any other entry after a tool call
// gets breathing room.
func needsGap(prev *chatEntry, kind core.UpdateKind) bool {
	return prev != nil && prev.kind == core.UpdateToolCall && kind != core.UpdateToolCall
}
