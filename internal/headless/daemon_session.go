package headless

import (
	"context"
	"errors"

	"jin/internal/session"
	"jin/internal/store"
)

func daemonPromptSession(ctx context.Context, db *store.DB, dir string, opt Options) (session.Snapshot, error) {
	if opt.Session == "" && !opt.Continue {
		return daemonClient.Create(ctx, dir)
	}
	var rec store.Session
	var found bool
	var err error
	if opt.Continue {
		list, failure := db.ListByPath(dir)
		err = failure
		if len(list) > 0 {
			rec, found = list[0], true
		}
	} else {
		rec, found, err = db.GetSession(opt.Session)
		if err == nil && !found {
			rec, found, err = db.SessionByPrefix(opt.Session)
		}
	}
	if err != nil {
		return session.Snapshot{}, err
	}
	if !found {
		return session.Snapshot{}, errors.New("session not found")
	}
	return daemonClient.Open(ctx, rec.ID)
}
