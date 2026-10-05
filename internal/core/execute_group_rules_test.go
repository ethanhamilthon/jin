package core

import (
	"testing"

	"jin/internal/provider"
)

func groupSizes(calls []provider.ToolCall) []int {
	var sizes []int
	for start := 0; start < len(calls); {
		n := len(nextGroup(calls[start:]))
		sizes = append(sizes, n)
		start += n
	}
	return sizes
}

func TestGroupingRules(t *testing.T) {
	cases := []struct {
		name  string
		calls []provider.ToolCall
		want  []int
	}{
		{"bash and custom tools together", []provider.ToolCall{
			fakeCall("1", "bash", `{"command":"a"}`), fakeCall("2", "bash", `{"command":"b"}`), fakeCall("3", "read", `{"path":"x"}`)}, []int{3}},
		{"edits of different files together", []provider.ToolCall{
			fakeCall("1", "edit", `{"path":"a.go"}`), fakeCall("2", "write", `{"path":"b.go"}`)}, []int{2}},
		{"second change of one file waits", []provider.ToolCall{
			fakeCall("1", "edit", `{"path":"a.go"}`), fakeCall("2", "edit", `{"path":"./a.go"}`)}, []int{1, 1}},
		{"read after a change of the same file waits", []provider.ToolCall{
			fakeCall("1", "write", `{"path":"a.go"}`), fakeCall("2", "read", `{"path":"a.go"}`)}, []int{1, 1}},
		{"change after a read of the same file waits", []provider.ToolCall{
			fakeCall("1", "read", `{"path":"a.go"}`), fakeCall("2", "edit", `{"path":"a.go"}`)}, []int{1, 1}},
		{"reads of one file together", []provider.ToolCall{
			fakeCall("1", "read", `{"path":"a.go"}`), fakeCall("2", "read", `{"path":"a.go"}`)}, []int{2}},
		{"ask_user and todo updates run alone", []provider.ToolCall{
			fakeCall("1", "bash", `{}`), fakeCall("2", "ask_user", `{}`), fakeCall("3", "todo", `{"items":[{"text":"x"}]}`), fakeCall("4", "todo", `{}`), fakeCall("5", "bash", `{}`)}, []int{1, 1, 1, 2}},
	}
	for _, c := range cases {
		got := groupSizes(c.calls)
		if len(got) != len(c.want) {
			t.Errorf("%s: groups %v, want %v", c.name, got, c.want)
			continue
		}
		for i := range got {
			if got[i] != c.want[i] {
				t.Errorf("%s: groups %v, want %v", c.name, got, c.want)
				break
			}
		}
	}
}
