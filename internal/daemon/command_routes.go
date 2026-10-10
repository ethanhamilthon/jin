package daemon

import (
	"encoding/json"
	"net/http"
	"sync"

	"jin/internal/session"
)

type commandCache struct {
	mu      sync.Mutex
	results map[string]*commandResult
}

func commandRead(action string) bool {
	switch action {
	case "snapshot", "live", "points", "context", "tasks":
		return true
	}
	return false
}

type commandResult struct {
	done   chan struct{}
	result Result
}

func commandRoute(version string, manager *session.Manager) http.HandlerFunc {
	cache := &commandCache{results: map[string]*commandResult{}}
	return func(w http.ResponseWriter, r *http.Request) {
		var command Command
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 12<<20)).Decode(&command); err != nil {
			reply(w, 400, map[string]string{"error": err.Error()})
			return
		}
		if command.Version != version {
			reply(w, 409, map[string]string{"error": "client and daemon versions differ"})
			return
		}
		if command.ID == "" {
			reply(w, 400, map[string]string{"error": "command id is required"})
			return
		}
		if commandRead(command.Action) {
			value, err := execute(r.Context(), manager, command)
			if err != nil {
				reply(w, 200, Result{Error: err.Error()})
				return
			}
			encoded, err := json.Marshal(value)
			if err != nil {
				reply(w, 200, Result{Error: err.Error()})
				return
			}
			reply(w, 200, Result{Value: encoded})
			return
		}
		cache.mu.Lock()
		if cached, ok := cache.results[command.ID]; ok {
			cache.mu.Unlock()
			select {
			case <-cached.done:
				reply(w, 200, cached.result)
			case <-r.Context().Done():
			}
			return
		}
		if len(cache.results) >= 100000 {
			cache.mu.Unlock()
			reply(w, 503, map[string]string{"error": "command cache is full; restart the daemon when idle"})
			return
		}
		cached := &commandResult{done: make(chan struct{})}
		cache.results[command.ID] = cached
		cache.mu.Unlock()
		value, err := execute(r.Context(), manager, command)
		result := Result{}
		if err != nil {
			result.Error = err.Error()
		} else {
			result.Value, err = json.Marshal(value)
			if err != nil {
				result.Error = err.Error()
			}
		}
		cached.result = result
		close(cached.done)
		reply(w, 200, result)
	}
}
