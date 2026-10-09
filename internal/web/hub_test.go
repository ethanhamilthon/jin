package web

import "testing"

func TestSlowPageGetsResyncInsteadOfDrop(t *testing.T) {
	h := newHub()
	c := make(chan []byte, 4)
	h.clients[c] = ""
	for i := 0; i < 10; i++ {
		h.publish(map[string]int{"n": i})
	}
	if _, ok := h.clients[c]; !ok {
		t.Fatal("slow page was dropped")
	}
	if got := string(<-c); got != string(resyncEvent) {
		t.Fatalf("first queued event = %s, want resync", got)
	}
}
