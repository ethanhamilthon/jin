package session

import "sync"

type eventBus struct {
	mu      sync.Mutex
	clients map[chan Event]bool
}

// Subscribe returns an independent stream. Overflow requires a fresh snapshot.
func (m *Manager) Subscribe() (<-chan Event, func()) {
	bus := &m.events
	bus.mu.Lock()
	if bus.clients == nil {
		bus.clients = map[chan Event]bool{}
	}
	stream := make(chan Event, 256)
	bus.clients[stream] = true
	bus.mu.Unlock()
	return stream, func() {
		bus.mu.Lock()
		defer bus.mu.Unlock()
		if bus.clients[stream] {
			delete(bus.clients, stream)
			close(stream)
		}
	}
}

func drainEvents(stream chan Event) {
	for {
		select {
		case <-stream:
		default:
			return
		}
	}
}

func (b *eventBus) publish(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for stream := range b.clients {
		select {
		case stream <- event:
		default:
			drainEvents(stream)
			stream <- Event{Type: "resync", Seq: event.Seq}
		}
	}
}
