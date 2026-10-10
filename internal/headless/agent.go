package headless

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"

	"jin/internal/agentkit"
	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/sources"
	"jin/internal/startup"
	"jin/internal/store"
)

// buildAgent runs the commands of the system prompt file and the hooks,
// then makes the agent. The #prompt bodies are rendered only when the prompt
// has a # in it; they come back for prompts.Expand.
func buildAgent(ctx context.Context, db *store.DB, dir, id, prompt string, names []string, cfg store.Config, save bool, out writer) (*core.Agent, map[string]string) {
	input := startup.Input{Dir: dir, SessionID: id, PromptsDisabled: cfg.PromptsDisabled, WithPrompts: strings.Contains(prompt, "#")}
	rendered := startup.Render(ctx, input, nil)
	for _, warning := range rendered.Warnings {
		out.Progress("jin: " + warning)
	}
	client := sources.FromConfig(db, cfg)
	client.SetStallTimeout(cfg.StallTimeout)
	owner := id
	if owner == "" {
		owner = "headless"
	}
	agent := agentkit.New(agentkit.Spec{
		Mode: agentkit.OneShot, Client: client, Names: names, Dir: dir, Owner: owner,
	})
	agentkit.Apply(agent, rendered)
	agent.SetRefresher(agentkit.Refresher(input))
	return agent, rendered.Prompts
}

// openSession finds the session to continue, if any, and its messages.
func openSession(db *store.DB, opt Options, dir string) (store.Session, []provider.Message, error) {
	var record store.Session
	switch {
	case opt.Session != "":
		found, ok, err := db.GetSession(opt.Session)
		if err != nil {
			return record, nil, err
		}
		if !ok {
			return record, nil, fmt.Errorf("no session with id %q", opt.Session)
		}
		record = found
	case opt.Continue:
		list, err := db.ListByPath(dir)
		if err != nil {
			return record, nil, err
		}
		if len(list) == 0 {
			return record, nil, errors.New("no session to continue in this directory")
		}
		record = list[0]
	default:
		return record, nil, nil
	}
	messages, err := db.LoadMessages(record.ID)
	return record, messages, err
}

func newID() string {
	var buf [16]byte
	rand.Read(buf[:])
	return hex.EncodeToString(buf[:])
}

func firstLine(text string) string {
	line, _, _ := strings.Cut(strings.TrimSpace(text), "\n")
	return line
}

func truncate(text string, limit int) string {
	runes := []rune(text)
	if len(runes) <= limit {
		return text
	}
	return string(runes[:limit-1]) + "…"
}
