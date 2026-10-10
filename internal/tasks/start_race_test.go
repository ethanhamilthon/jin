package tasks

import "testing"

func TestFastTaskPublishesInitializedInfo(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	manager := New()
	defer manager.StopAll()
	for range 20 {
		info, err := manager.Start("owner", ":", t.TempDir(), true)
		if err != nil {
			t.Fatal(err)
		}
		if !info.Stdin || info.Status != Running || info.Exit != -1 {
			t.Fatalf("initial info: %+v", info)
		}
		event := waitEvent(t, manager)
		if !event.Task.Stdin || event.Task.ID != info.ID || event.Task.Status != Done {
			t.Fatalf("event: %+v", event)
		}
	}
}
