package ui

import (
	"os"
	"testing"

	"jin/internal/provider"
)

func TestRefreshedAsyncMessageAcknowledgesItsEventExactlyOnce(t *testing.T) {
	a, s := asyncApp(t)
	a.receiveAsync(claimedBatch(t, a, os.Getpid()))
	message := provider.Message{Role: "user", Content: "<system-refreshed>instructions were refreshed</system-refreshed>\n\n" + taskResultText}
	s.persistMessage(message)
	if len(s.asyncAcks) != 0 {
		t.Fatal("refreshed async message left its acknowledgement pending")
	}
	a.retryAsync(s.id)
	if count := eventCount(t, a, os.Getpid()); count != 0 {
		t.Fatalf("refreshed async result can replay: %d events", count)
	}
	messages, err := a.store.LoadMessages(s.id)
	if err != nil || len(messages) != 1 || messages[0].Content != message.Content {
		t.Fatalf("refreshed history not saved atomically: %+v, %v", messages, err)
	}
}
