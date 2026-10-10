package cliproxy

import (
	"encoding/json"
	"net/http"
)

type command struct {
	Action   string `json:"action"`
	Profile  string `json:"profile,omitempty"`
	State    string `json:"state,omitempty"`
	Version  string `json:"version,omitempty"`
	Disabled bool   `json:"disabled,omitempty"`
}

func (b *broker) command(w http.ResponseWriter, r *http.Request) {
	var input command
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&input); err != nil {
		http.Error(w, "invalid command", 400)
		return
	}
	b.mu.Lock()
	if b.closing {
		b.mu.Unlock()
		http.Error(w, "proxy is stopping", 503)
		return
	}
	b.operations++
	b.mu.Unlock()
	defer func() { b.mu.Lock(); b.operations--; b.mu.Unlock(); b.notify() }()
	b.operation.Lock()
	defer b.operation.Unlock()
	if input.Action != "update" && input.Action != "status" {
		if err := b.restartIfNeeded(r.Context()); err != nil {
			http.Error(w, err.Error(), 503)
			return
		}
	}
	var value any
	var err error
	switch input.Action {
	case "accounts":
		value, err = b.accounts(r.Context())
	case "login":
		value, err = b.login(r.Context(), input.Profile)
	case "poll":
		value, err = b.poll(r.Context(), input.Profile, input.State)
	case "cancel":
		value, err = b.cancelLogin(r.Context(), input.State)
	case "logout":
		err = b.logout(r.Context(), input.Profile)
		value = map[string]bool{"ok": err == nil}
	case "sync":
		err = b.syncProfile(r.Context(), input.Profile, input.Disabled)
		value = map[string]bool{"ok": err == nil}
	case "update":
		err = b.update(r.Context(), input.Version)
		value = Installed(b.root)
	case "status":
		value = Installed(b.root)
	default:
		http.Error(w, "unknown operation", 400)
		return
	}
	if err != nil {
		http.Error(w, err.Error(), 409)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(value)
}
