package prompts

import (
	"os"
	"strings"
	"testing"
)

func setup(t *testing.T, files map[string]string) {
	t.Setenv("HOME", t.TempDir())
	for name, body := range files {
		path, err := Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestExpandWrapsReferencedPrompts(t *testing.T) {
	setup(t, map[string]string{"review/security": "Check for injection.\n", "style": "Be terse."})
	got := Expand("hey, use #review/security. and #nope #style #review/security", Bodies(nil))
	for _, want := range []string{
		"<pasted-prompts>",
		"<prompt name=\"review/security\">\nCheck for injection.\n</prompt>",
		"<prompt name=\"style\">\nBe terse.\n</prompt>",
		"</pasted-prompts>\n\nhey, use #review/security.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in\n%s", want, got)
		}
	}
	if strings.Count(got, "<prompt name=\"review/security\">") != 1 {
		t.Error("a prompt referenced twice must be included once")
	}
}

func TestExpandLeavesUnknownReferencesAlone(t *testing.T) {
	setup(t, map[string]string{"style": "Be terse."})
	for _, text := range []string{"fix #123", "# Title", "a#style", "plain"} {
		if got := Expand(text, Bodies(nil)); got != text {
			t.Errorf("Expand(%q) = %q, want unchanged", text, got)
		}
	}
}

func TestStripRestoresTypedText(t *testing.T) {
	setup(t, map[string]string{"style": "Be terse."})
	text := "use #style please"
	expanded := Expand(text, Bodies(nil))
	if expanded == text {
		t.Fatal("expected expansion")
	}
	if got := Strip(expanded); got != text {
		t.Errorf("Strip = %q, want %q", got, text)
	}
}

func TestPathRejectsEscapes(t *testing.T) {
	for _, name := range []string{"", "../x", "a/../b", "/abs", "a//b", ".hidden", "a b"} {
		if _, err := Path(name); err == nil {
			t.Errorf("Path(%q) should fail", name)
		}
	}
}

func TestListAndDelete(t *testing.T) {
	setup(t, map[string]string{"a/b/c": "x", "d": "y"})
	names, err := List()
	if err != nil || strings.Join(names, ",") != "plan,review,a/b/c,d" {
		t.Fatalf("List = %v, %v", names, err)
	}
	if err := Delete("a/b/c"); err != nil {
		t.Fatal(err)
	}
	names, _ = List()
	if strings.Join(names, ",") != "plan,review,d" {
		t.Errorf("after delete: %v", names)
	}
}

func TestExpandNamesUsesOnlyTheGivenNames(t *testing.T) {
	bodies := map[string]string{"plan": "Plan body.", "style": "Style body.", "empty": " "}
	got := ExpandNames("use #style and #plan", []string{"plan", "empty", "nope"}, bodies)
	if !strings.Contains(got, "Plan body.") || strings.Contains(got, "Style body.") || strings.Contains(got, `name="empty"`) {
		t.Fatalf("got:\n%s", got)
	}
	if !strings.HasSuffix(got, "\n\nuse #style and #plan") {
		t.Fatalf("the text must follow unchanged:\n%s", got)
	}
	if got := ExpandNames("plain", nil, bodies); got != "plain" {
		t.Fatalf("no names must leave the text alone, got %q", got)
	}
}
