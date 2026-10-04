package core

import (
	"slices"
	"strconv"
	"strings"

	"jin/internal/provider"
)

const omittedPrefix = "[output of "

// pruneToolResults returns a copy of history in which every tool result
// before the last keep turns is replaced by a short note. A turn starts with
// a user prompt. Tool calls and their results stay paired; only the text of
// a result changes. changed is false when nothing was worth replacing.
func pruneToolResults(history []provider.Message, keep int) (pruned []provider.Message, changed bool) {
	cut := turnStart(history, keep)
	names := make(map[string]string)
	pruned = slices.Clone(history)
	for i, msg := range pruned[:cut] {
		for _, call := range msg.ToolCalls {
			names[call.ID] = call.Function.Name
		}
		if msg.Role != "tool" || strings.HasPrefix(msg.Content, omittedPrefix) {
			continue
		}
		note := omittedNote(names[msg.ToolCallID], len(msg.Content))
		if len(note) < len(msg.Content) {
			pruned[i].Content, changed = note, true
		}
	}
	return pruned, changed
}

func omittedNote(tool string, size int) string {
	if tool == "" {
		tool = "tool"
	}
	return omittedPrefix + tool + " omitted, " + strconv.Itoa(size) + " bytes; run it again if needed]"
}

// turnStart is the index of the prompt that opens the keep-th last turn, or
// 0 when the history has no more than keep turns.
func turnStart(history []provider.Message, keep int) int {
	for i := len(history) - 1; i >= 0 && keep > 0; i-- {
		if history[i].Role == "user" && history[i].Content != imagesFromTools {
			if keep--; keep == 0 {
				return i
			}
		}
	}
	return 0
}

// prune shortens old tool results in the in-memory history and lowers the
// size estimate by what it removed. The stored history keeps the full text.
func (a *Agent) prune(history *[]provider.Message, keep int) bool {
	pruned, changed := pruneToolResults(*history, keep)
	if !changed {
		return false
	}
	freed := 0
	for i := range min(a.mark, len(pruned)) {
		freed += len((*history)[i].Content) - len(pruned[i].Content)
	}
	a.size = max(0, a.size-freed/4)
	*history = pruned
	return true
}

// needsPrune is true once the context fills 70% of the model's window.
func needsPrune(size, window int) bool {
	return window > 0 && size*10 >= window*7
}

// pruneKeepTurns is how many recent turns keep their full tool output.
const pruneKeepTurns = 4
