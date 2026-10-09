package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// hub sends every event to every open page as Server-Sent Events. A page
// too slow to keep up loses its queued events and gets a resync event.
type hub struct {
	mu      sync.Mutex
	clients map[chan []byte]string
}

func newHub() *hub { return &hub{clients: map[chan []byte]string{}} }

func (h *hub) publish(event any) {
	data, err := json.Marshal(event)
	if err != nil {
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		select {
		case c <- data:
		default:
			overflow(c)
		}
	}
}

// overflow drops what a slow page missed and tells it to load its state
// again. The stream stays open, so the page never shows a lost connection.
func overflow(c chan []byte) {
	for {
		select {
		case <-c:
		default:
			c <- resyncEvent
			return
		}
	}
}

var resyncEvent = []byte(`{"type":"resync"}`)

func (h *hub) serve(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	c := make(chan []byte, 1024)
	device := deviceID(r)
	h.mu.Lock()
	h.clients[c] = device
	h.mu.Unlock()
	h.publish(devicesEvent)
	defer h.drop(c)
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	fmt.Fprint(w, "retry: 1000\n\n")
	flusher.Flush()
	ping := time.NewTicker(20 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case data, open := <-c:
			if !open {
				return
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-ping.C:
			fmt.Fprint(w, ": ping\n\n")
			flusher.Flush()
		}
	}
}

func (h *hub) drop(c chan []byte) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c)
	}
	h.mu.Unlock()
	h.publish(devicesEvent)
}

func (h *hub) closeAll() {
	h.mu.Lock()
	defer h.mu.Unlock()
	for c := range h.clients {
		delete(h.clients, c)
		close(c)
	}
}
