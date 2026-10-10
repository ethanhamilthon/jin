package sources

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"jin/internal/store"
)

func Revision(db *store.DB) string {
	if db == nil {
		return ""
	}
	providers, _, err := db.LoadProviders()
	if err != nil {
		return ""
	}
	snapshot := map[string]any{"providers": providers}
	for _, p := range providers {
		scope, _ := db.LoadScopeFor(p.ID)
		snapshot[p.ID] = scope
	}
	data, _ := json.Marshal(snapshot)
	digest := sha256.Sum256(data)
	return fmt.Sprintf("%x", digest[:])
}
