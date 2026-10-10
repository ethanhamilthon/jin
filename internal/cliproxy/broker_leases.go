package cliproxy

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"time"
)

func (b *broker) register(w http.ResponseWriter, r *http.Request) {
	var data [16]byte
	if _, err := rand.Read(data[:]); err != nil {
		http.Error(w, "cannot register", 500)
		return
	}
	id := hex.EncodeToString(data[:])
	b.mu.Lock()
	if b.closing {
		b.mu.Unlock()
		http.Error(w, "proxy is stopping", 503)
		return
	}
	b.clients[id] = true
	b.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"client": id})
	w.(http.Flusher).Flush()
	<-r.Context().Done()
	b.mu.Lock()
	delete(b.clients, id)
	b.mu.Unlock()
	b.notify()
}

func (b *broker) lease(w http.ResponseWriter, r *http.Request) {
	if err := b.recoverForLease(r.Context()); err != nil {
		http.Error(w, err.Error(), 503)
		return
	}
	for {
		b.mu.Lock()
		known := b.clients[r.URL.Query().Get("client")]
		if !known || b.closing {
			b.mu.Unlock()
			http.Error(w, "proxy client disconnected", 409)
			return
		}
		if !b.updating {
			b.active++
			break
		}
		b.mu.Unlock()
		select {
		case <-r.Context().Done():
			return
		case <-time.After(50 * time.Millisecond):
		}
	}
	p := b.proxy
	b.mu.Unlock()
	defer func() { b.mu.Lock(); b.active--; b.mu.Unlock(); b.notify() }()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"endpoint": p.endpoint + "/v1", "key": p.keys.Client})
	w.(http.Flusher).Flush()
	<-r.Context().Done()
}

func (b *broker) waitIdle(ctx context.Context) error {
	for {
		b.mu.Lock()
		idle := b.active == 0
		b.mu.Unlock()
		if idle {
			return nil
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
		}
	}
}
