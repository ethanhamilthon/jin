package session

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/sources"
	"jin/internal/store"
)

// UserTurns counts the messages the user wrote, without the background tasks notes.
func UserTurns(messages []provider.Message) int {
	n := 0
	for _, msg := range messages {
		if msg.Role == "user" && !core.IsTasksNote(msg.Content) {
			n++
		}
	}
	return n
}

// TitleDue says whether a session with this many user messages is named now.
// last is the count at which it was named before, so one count names it once.
func TitleDue(t store.TitleSettings, turns, last int) bool {
	if t.After == 0 || turns == last {
		return false
	}
	return turns == t.After || (t.Refresh && turns == store.TitleRefreshTurns)
}

// NameSession asks the title model, or the session's own model when none is
// set, names the stored session after its history and saves the name.
func NameSession(ctx context.Context, db *store.DB, cfg store.Config, tools json.RawMessage, id, providerID, model string) (string, error) {
	if cfg.Title.Model != "" {
		providerID, model = cfg.Title.Provider, cfg.Title.Model
	}
	client, _, missing := sources.ClientFor(db, cfg, providerID)
	if missing {
		return "", errors.New("no provider " + providerID)
	}
	messages, err := db.LoadMessages(id)
	if err != nil {
		return "", err
	}
	answer, err := core.Title(ctx, client, model, cfg.Title.Effort, core.SinceLastSummary(messages), tools, cfg.Title.Prompt)
	if err != nil {
		return "", err
	}
	title := cleanTitle(answer)
	if title == "" {
		return "", errors.New("the model returned no title")
	}
	return title, db.SetTitle(id, title)
}

// cleanTitle keeps the first line of an answer, without the quotes and
// full stop a model may put around a title, and cuts it to maxTitle.
func cleanTitle(answer string) string {
	return Cut(strings.Trim(FirstLine(answer), " \t\"'`*.“”‘’"), maxTitle)
}
