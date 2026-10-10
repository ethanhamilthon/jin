package session

import (
	"context"
	"sync"
	"testing"
)

func TestIndependentSubscribers(t *testing.T) {
	manager := NewManager(context.Background(), nil, "test", nil)
	first, closeFirst := manager.Subscribe()
	second, closeSecond := manager.Subscribe()
	defer closeSecond()
	manager.publish(Event{Type: "config"})
	for _, stream := range []<-chan Event{first, second} {
		if event := <-stream; event.Type != "config" || event.Seq != 1 {
			t.Fatalf("event: %+v", event)
		}
	}
	closeFirst()
	closeFirst()
	manager.publish(Event{Type: "tasks"})
	if _, open := <-first; open {
		t.Fatal("subscription remains open")
	}
	if event := <-second; event.Type != "tasks" {
		t.Fatalf("event: %+v", event)
	}
}

func TestSlowSubscriberResync(t *testing.T) {
	manager := NewManager(context.Background(), nil, "test", nil)
	stream, closeStream := manager.Subscribe()
	defer closeStream()
	for i := 0; i < 257; i++ {
		manager.publish(Event{Type: "delta"})
	}
	if event := <-stream; event.Type != "resync" || event.Seq != 257 {
		t.Fatalf("event: %+v", event)
	}
}

func TestConcurrentSubscriptionClose(t *testing.T) {
	manager := NewManager(context.Background(), nil, "test", nil)
	var wait sync.WaitGroup
	for i := 0; i < 4; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for n := 0; n < 100; n++ {
				_, closeStream := manager.Subscribe()
				manager.publish(Event{Type: "config"})
				closeStream()
			}
		}()
	}
	wait.Wait()
}
