package tools

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func schemaOf(t *testing.T, tool Tool) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal(tool.Schema(), &out); err != nil {
		t.Fatalf("%s schema is not JSON: %v", tool.Name(), err)
	}
	return out["function"].(map[string]any)
}

func property(fn map[string]any, name string) map[string]any {
	return fn["parameters"].(map[string]any)["properties"].(map[string]any)[name].(map[string]any)
}

func TestEverySchemaIsValidJSON(t *testing.T) {
	for _, name := range Catalog() {
		tool, _ := Build([]string{name}, &MemoryTodos{}).Get(name)
		if schemaOf(t, tool)["description"] == "" {
			t.Errorf("%s has no description", name)
		}
	}
}

func TestBashDescriptionFollowsMode(t *testing.T) {
	tui := schemaOf(t, NewBash())["description"].(string)
	headless := schemaOf(t, NewBashHeadless())["description"].(string)
	if !strings.Contains(tui, "moves to the background") || strings.Contains(tui, "killed") {
		t.Errorf("tui = %q", tui)
	}
	if !strings.Contains(headless, "is killed") || strings.Contains(headless, "background") {
		t.Errorf("headless = %q", headless)
	}
	for _, d := range []string{tui, headless} {
		for _, fact := range []string{"new shell", "No TTY and no stdin", "16 KB"} {
			if !strings.Contains(d, fact) {
				t.Errorf("missing %q in %q", fact, d)
			}
		}
	}
	if _, ok := BuildHeadless([]string{"bash"}, nil).Get("bash"); !ok {
		t.Fatal("headless registry has no bash")
	}
}

func TestSchemaMinimums(t *testing.T) {
	for _, tc := range []struct {
		tool Tool
		name string
	}{{NewBash(), "timeout"}, {NewBashHeadless(), "timeout"}, {NewRead(), "offset"}, {NewRead(), "limit"}} {
		if property(schemaOf(t, tc.tool), tc.name)["minimum"] != float64(1) {
			t.Errorf("%s.%s has no minimum 1", tc.tool.Name(), tc.name)
		}
	}
	questions := property(schemaOf(t, NewAsk()), "questions")
	if questions["minItems"] != float64(1) {
		t.Errorf("questions = %v", questions)
	}
}

func TestLocalValidationOfNumbers(t *testing.T) {
	file := writeTempFile(t, "a\nb\n")
	tests := []struct {
		tool Tool
		args string
		want string
	}{
		{NewBash(), `{"command":"true","timeout":0}`, "timeout must be at least 1"},
		{NewRead(), `{"path":"` + file + `","offset":0}`, "offset must be at least 1"},
		{NewRead(), `{"path":"` + file + `","limit":-2}`, "limit must be at least 1"},
	}
	for _, tt := range tests {
		_, err := tt.tool.Run(context.Background(), tt.args)
		if err == nil || !strings.Contains(err.Error(), tt.want) {
			t.Errorf("%s %s: err = %v", tt.tool.Name(), tt.args, err)
		}
	}
	if _, err := NewRead().Run(context.Background(), `{"path":"`+file+`","offset":2,"limit":1}`); err != nil {
		t.Fatal(err)
	}
}

func TestToolDescriptionsStateTheRules(t *testing.T) {
	read := schemaOf(t, NewRead())["description"].(string)
	edit := schemaOf(t, NewEdit())["description"].(string)
	write := schemaOf(t, NewWrite())["description"].(string)
	todo := schemaOf(t, NewTodo(&MemoryTodos{}))["description"].(string)
	for text, parts := range map[string][]string{
		read:  {"not part of the file", "32 KB"},
		edit:  {"exactly once", "line-number prefix", "Read the file first"},
		write: {"Read an existing file first"},
		todo:  {"replaces", "as soon as it is finished"},
	} {
		for _, p := range parts {
			if !strings.Contains(text, p) {
				t.Errorf("missing %q in %q", p, text)
			}
		}
	}
}
