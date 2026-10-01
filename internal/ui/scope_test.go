package ui

import (
	"slices"
	"testing"
)

func TestFilterScope(t *testing.T) {
	all := []string{"a", "b", "c"}
	cases := []struct {
		scope, want []string
	}{
		{nil, all},
		{[]string{"b"}, []string{"b"}},
		{[]string{"c", "a"}, []string{"a", "c"}},
		{[]string{"gone"}, all},
	}
	for _, c := range cases {
		if got := filterScope(all, c.scope); !slices.Equal(got, c.want) {
			t.Errorf("filterScope(%v) = %v, want %v", c.scope, got, c.want)
		}
	}
}

func TestToggleScope(t *testing.T) {
	all := []string{"a", "b", "c"}
	scope := toggleScope(all, nil, "b")
	if !slices.Equal(scope, []string{"a", "c"}) {
		t.Fatalf("disable from all = %v", scope)
	}
	if scope = toggleScope(all, scope, "b"); scope != nil {
		t.Fatalf("enabling the last missing model should clear the scope, got %v", scope)
	}
	one := []string{"a"}
	if got := toggleScope(all, one, "a"); !slices.Equal(got, one) {
		t.Errorf("the last enabled model must stay on, got %v", got)
	}
}

func TestNextModel(t *testing.T) {
	models := []string{"a", "b", "c"}
	for current, want := range map[string]string{"a": "b", "c": "a", "x": "a"} {
		if got := nextModel(models, current); got != want {
			t.Errorf("nextModel(%q) = %q, want %q", current, got, want)
		}
	}
	if nextModel(nil, "a") != "" {
		t.Error("empty list has no next model")
	}
}
