package store

import (
	"sync"
	"testing"

	"jin/internal/provider"
)

func TestNewSessionHoldsOneTitleClaim(t *testing.T) {
	db, _ := openTwo(t)
	created, err := db.TouchFirst("s1", "", "m", "", "fallback", "")
	if err != nil || !created {
		t.Fatalf("first touch = %v, %v", created, err)
	}
	if created, _ = db.TouchFirst("s1", "", "m", "", "other", ""); created {
		t.Fatal("second touch created the session again")
	}
	var wins int
	var mu sync.Mutex
	var wg sync.WaitGroup
	for range 12 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			applied, err := db.SetTitle("s1", "Named")
			if err != nil {
				t.Error(err)
			}
			mu.Lock()
			if applied {
				wins++
			}
			mu.Unlock()
		}()
	}
	wg.Wait()
	if wins != 1 {
		t.Fatalf("titles applied = %d, want 1", wins)
	}
}

func TestOldSessionIsNotRenamed(t *testing.T) {
	db, _ := openTwo(t)
	if _, err := db.sql.Exec(`INSERT INTO sessions(id,path,model,effort,title,created_at,updated_at,provider) VALUES('old','','m','','Old',1,1,'')`); err != nil {
		t.Fatal(err)
	}
	if applied, err := db.SetTitle("old", "New"); err != nil || applied {
		t.Fatalf("old session renamed: %v %v", applied, err)
	}
	if applied, err := db.SetTitle("missing", "New"); err != nil || applied {
		t.Fatalf("missing session: %v %v", applied, err)
	}
}

func TestSetTitle(t *testing.T) {
	db, _ := openTwo(t)
	if err := db.TouchProvider("s1", "", "m", "", "first prompt", ""); err != nil {
		t.Fatal(err)
	}
	for _, text := range []string{"a", "b"} {
		if err := db.AppendMessage("s1", provider.Message{Role: "user", Content: text}); err != nil {
			t.Fatal(err)
		}
	}
	if applied, err := db.SetTitle("s1", "Renamed"); err != nil || !applied {
		t.Fatal(err)
	}
	rec, ok, err := db.GetSession("s1")
	if err != nil || !ok || rec.Title != "Renamed" {
		t.Fatalf("session = %+v %v %v", rec, ok, err)
	}
}
