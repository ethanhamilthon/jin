package session

import (
	"context"
	"errors"
	"strings"

	"jin/internal/core"
	"jin/internal/sources"
	"jin/internal/store"
)

// NamePrompt names the immutable first prompt with the captured provider settings.
func NamePrompt(ctx context.Context, db *store.DB, cfg store.Config, id, providerID, model, effort, prompt string) (string, error) {
	client, _, missing := sources.ClientFor(db, cfg, providerID)
	if missing {
		return "", errors.New("no provider " + providerID)
	}
	answer, err := core.Title(ctx, client, model, effort, prompt)
	if err != nil {
		return "", err
	}
	title := cleanTitle(answer)
	if title == "" {
		return "", errors.New("the model returned no title")
	}
	if !plausibleTitle(answer, title) {
		return "", nil
	}
	applied, err := db.SetTitle(id, title)
	if err != nil || !applied {
		return "", err
	}
	return title, nil
}

// cleanTitle keeps the first line of an answer, without the quotes and
// full stop a model may put around a title, and cuts it to maxTitle.
func cleanTitle(answer string) string {
	return Cut(strings.Trim(FirstLine(answer), " \t\"'`*.“”‘’"), maxTitle)
}

const maxTitleWords = 8

// plausibleTitle refuses an answer that is a reply to the subject, not a title:
// several lines, too many words, or text that had to be cut.
func plausibleTitle(answer, title string) bool {
	return !strings.Contains(strings.TrimSpace(answer), "\n") && len(strings.Fields(title)) <= maxTitleWords && !strings.HasSuffix(title, "…")
}
