package cliproxy

import (
	"encoding/json"
	"net/http"
	"strings"
)

func fakeManagement(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/v0/management")
	value := map[string]any{"status": "ok"}
	switch {
	case path == "/force-model-prefix":
		value = map[string]any{"force-model-prefix": true}
	case path == "/auth-files":
		value = map[string]any{"files": []any{}}
	case strings.HasSuffix(path, "-auth-url"):
		profile := strings.TrimSuffix(strings.TrimPrefix(path, "/"), "-auth-url")
		hosts := map[string]string{"anthropic": "claude.ai", "codex": "auth.openai.com", "antigravity": "accounts.google.com"}
		value = map[string]any{"state": "fixture-state-" + profile, "url": "https://" + hosts[profile] + "/oauth/authorize", "status": "ok"}
	case path == "/get-auth-status":
		value = map[string]any{"status": "wait"}
	case path == "/oauth-session":
		value = map[string]any{"status": "ok", "cancelled": true}
	}
	json.NewEncoder(w).Encode(value)
}
