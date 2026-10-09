package ui

import (
	"jin/internal/core"
	"jin/internal/provider"
)

const largestResults = 5

// showContext prints what fills the context of the open session.
func (a *app) showContext(full bool) {
	s := a.active
	if !s.ready {
		s.appendEntry(chatEntry{kind: core.UpdateInfo, text: "The session is still starting"})
		return
	}
	messages, err := s.store.LoadMessages(s.id)
	if err != nil {
		a.report(err)
		return
	}
	info := contextInfo{
		used: s.usage.Context, window: s.window(), cache: s.cache,
		prompt:      core.ExplainPrompt(s.agent.SystemPrompt()),
		toolSchemas: s.agent.ToolSchemaBytes(), tools: s.agent.ToolSchemaParts(), full: full,
	}
	base := promptBytes(info.prompt) + info.toolSchemas
	messages = core.EffectiveMessages(core.SinceLastSummary(messages), base, info.used, info.window)
	info.messages, info.conversation = len(messages), core.ConversationBytes(messages)
	for _, r := range core.LargestToolResults(messages, largestResults) {
		info.results = append(info.results, contextResult{label: a.callLabel(r.Call), bytes: r.Bytes})
	}
	s.closeOpenEntry()
	s.appendEntry(chatEntry{kind: core.UpdateInfo, text: info.String()})
}

func (a *app) callLabel(call provider.ToolCall) string {
	return call.Function.Name + " " + truncate(firstLine(toolSummary(a.registry, call)), 60)
}
