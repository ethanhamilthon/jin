package core

import (
	"encoding/json"
	"path/filepath"

	"jin/internal/provider"
	"jin/internal/tools"
)

type callOutcome struct {
	result  string
	images  []provider.Image
	changes []tools.Change
}

// exclusive reports whether the call must run alone: it asks the user or
// replaces the todo list.
func exclusive(call provider.ToolCall) bool {
	switch call.Function.Name {
	case "ask_user":
		return true
	case "todo":
		var args struct {
			Items json.RawMessage `json:"items"`
		}
		return json.Unmarshal([]byte(call.Function.Arguments), &args) != nil || len(args.Items) != 0 && string(args.Items) != "null"
	}
	return false
}

// filePath is the file a read, edit or write call names.
func filePath(call provider.ToolCall) (string, bool) {
	switch call.Function.Name {
	case "read", "edit", "write":
		var args struct {
			Path string `json:"path"`
		}
		if json.Unmarshal([]byte(call.Function.Arguments), &args) == nil && args.Path != "" {
			return filepath.Clean(args.Path), true
		}
	}
	return "", false
}

// nextGroup is the longest run of calls at the front that may run at the same
// time. Calls sent together in one answer are independent by contract, so
// bash and other tools run side by side; an exclusive call runs alone, and a
// call on a file that an earlier call of the group edits, writes or reads
// together with a change starts a new group.
func nextGroup(calls []provider.ToolCall) []provider.ToolCall {
	if exclusive(calls[0]) {
		return calls[:1]
	}
	reads, changes := map[string]bool{}, map[string]bool{}
	end := 0
	for ; end < len(calls) && !exclusive(calls[end]); end++ {
		path, ok := filePath(calls[end])
		if !ok {
			continue
		}
		changing := calls[end].Function.Name != "read"
		if changes[path] || changing && reads[path] {
			break
		}
		if changing {
			changes[path] = true
		} else {
			reads[path] = true
		}
	}
	return calls[:end]
}
