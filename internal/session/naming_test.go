package session

import (
	"testing"

	"jin/internal/provider"
	"jin/internal/store"
)

func TestTitleDue(t *testing.T) {
	cases := []struct {
		name        string
		set         store.TitleSettings
		turns, last int
		want        bool
	}{
		{"first message", store.TitleSettings{After: 1}, 1, 0, true},
		{"off", store.TitleSettings{After: 0, Refresh: true}, 4, 0, false},
		{"same count twice", store.TitleSettings{After: 1}, 1, 1, false},
		{"second message", store.TitleSettings{After: 1}, 2, 1, false},
		{"fourth without refresh", store.TitleSettings{After: 1}, 4, 1, false},
		{"fourth with refresh", store.TitleSettings{After: 1, Refresh: true}, 4, 1, true},
		{"after and refresh at four", store.TitleSettings{After: 4, Refresh: true}, 4, 0, true},
	}
	for _, c := range cases {
		if got := TitleDue(c.set, c.turns, c.last); got != c.want {
			t.Errorf("%s: got %v, want %v", c.name, got, c.want)
		}
	}
}

func TestUserTurnsSkipsOtherRoles(t *testing.T) {
	messages := []provider.Message{{Role: "user", Content: "a"}, {Role: "assistant", Content: "b"}, {Role: "tool", Content: "c"}, {Role: "user", Content: "d"}}
	if got := UserTurns(messages); got != 2 {
		t.Fatalf("turns = %d", got)
	}
}
