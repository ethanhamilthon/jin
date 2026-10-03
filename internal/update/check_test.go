package update

import (
	"strconv"
	"testing"
	"time"
)

type memorySettings map[string]string

func (m memorySettings) Setting(key string) (string, error) { return m[key], nil }
func (m memorySettings) SetSetting(key, value string) error {
	m[key] = value
	return nil
}

func TestAvailableUsesCache(t *testing.T) {
	releaseBase = "http://127.0.0.1:1"
	defer func() { releaseBase = "https://github.com" }()
	now := time.Unix(1_000_000, 0)
	db := memorySettings{keyLatest: "v0.7", keyChecked: strconv.FormatInt(now.Unix()-60, 10)}
	if got := Available(t.Context(), db, "v0.6.2", now); got != "v0.7" {
		t.Fatalf("cached = %q", got)
	}
	if got := Available(t.Context(), db, "v0.7", now); got != "" {
		t.Fatalf("same version = %q", got)
	}
	later := now.Add(7 * time.Hour)
	if got := Available(t.Context(), db, "v0.6.2", later); got != "v0.7" || db[keyChecked] != strconv.FormatInt(later.Unix(), 10) {
		t.Fatalf("offline recheck = %q %v", got, db)
	}
}
