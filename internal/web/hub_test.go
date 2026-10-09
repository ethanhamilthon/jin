package web

import "testing"

func TestSlowPageGetsResyncInsteadOfDrop(t *testing.T) {
	h := newHub()
	c := make(chan []byte, 4)
	h.clients[c] = true
	for i := 0; i < 10; i++ {
		h.publish(map[string]int{"n": i})
	}
	if !h.clients[c] {
		t.Fatal("slow page was dropped")
	}
	if got := string(<-c); got != string(resyncEvent) {
		t.Fatalf("first queued event = %s, want resync", got)
	}
}
