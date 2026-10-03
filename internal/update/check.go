package update

import (
	"context"
	"strconv"
	"time"
)

// Settings is the part of the store a background check needs.
type Settings interface {
	Setting(key string) (string, error)
	SetSetting(key, value string) error
}

const (
	keyLatest  = "update.latest"
	keyChecked = "update.checked"
	// checkEvery spaces the network checks; sessions in between reuse the
	// last answer from the settings.
	checkEvery = 6 * time.Hour
)

// Available returns a release newer than current, or "". It asks GitHub at
// most once per checkEvery and otherwise answers from the settings. Errors
// are quiet: an offline machine just sees no update.
func Available(ctx context.Context, db Settings, current string, now time.Time) string {
	latest, _ := db.Setting(keyLatest)
	checked, _ := db.Setting(keyChecked)
	seconds, _ := strconv.ParseInt(checked, 10, 64)
	if now.Sub(time.Unix(seconds, 0)) >= checkEvery {
		ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if tag, err := Latest(ctx); err == nil {
			latest = tag
			_ = db.SetSetting(keyLatest, tag)
		}
		_ = db.SetSetting(keyChecked, strconv.FormatInt(now.Unix(), 10))
	}
	if Newer(latest, current) {
		return latest
	}
	return ""
}
