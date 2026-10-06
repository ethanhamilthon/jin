package session

import (
	"errors"
	"strings"

	"jin/internal/core"
	"jin/internal/hooks"
	"jin/internal/store"
)

// Part is a named piece of the context with its approximate token count.
type Part struct {
	Name   string `json:"name"`
	Tokens int    `json:"tokens"`
}

// ContextReport is what fills the context of a session; token counts other
// than Used are estimates.
type ContextReport struct {
	Used         int    `json:"used"`
	Window       int    `json:"window"`
	Prompt       []Part `json:"prompt"`
	ToolSchemas  int    `json:"tool_schemas"`
	Messages     int    `json:"messages"`
	Conversation int    `json:"conversation"`
	Results      []Part `json:"results"`
	Cache        *int   `json:"cache,omitempty"`
}

// Context reports what fills the context of a session.
func (m *Manager) Context(id string) (ContextReport, error) {
	cfg, err := m.db.LoadConfig()
	if err != nil {
		return ContextReport{}, err
	}
	var report ContextReport
	err = m.Do(id, func(s *Session) error {
		if !s.ready {
			return errors.New("The session is still starting")
		}
		messages, err := m.db.LoadMessages(s.id)
		if err != nil {
			return err
		}
		prompt := core.ExplainPrompt(s.agent.SystemPrompt(), s.path, m.hookParts(cfg, s.path))
		report = ContextReport{Used: s.usage.Context, Window: s.window(), ToolSchemas: tokens(s.agent.ToolSchemaBytes()), Cache: s.cache}
		base := s.agent.ToolSchemaBytes()
		for _, part := range prompt {
			if part.Name != "cache break" {
				report.Prompt = append(report.Prompt, Part{part.Name, tokens(part.Bytes())})
				base += part.Bytes()
			}
		}
		messages = core.EffectiveMessages(core.SinceLastSummary(messages), base, report.Used, report.Window)
		report.Messages, report.Conversation = len(messages), tokens(core.ConversationBytes(messages))
		for _, r := range core.LargestToolResults(messages, 5) {
			label := r.Call.Function.Name + " " + Cut(FirstLine(ToolSummary(m.registry, r.Call)), 60)
			report.Results = append(report.Results, Part{label, tokens(r.Bytes)})
		}
		return nil
	})
	return report, err
}

func tokens(bytes int) int { return core.EstimateTokens(bytes) }

func (m *Manager) hookParts(cfg store.Config, dir string) []core.PromptPart {
	trust, _ := m.db.HooksTrust(dir)
	active, _ := hooks.LoadIn(dir, cfg.HooksDisabled, trust == store.Trusted)
	parts := make([]core.PromptPart, len(active))
	for i, hook := range active {
		parts[i] = core.PromptPart{Name: "hook " + hook.Name, Text: strings.TrimSpace(hook.Body)}
	}
	return parts
}
