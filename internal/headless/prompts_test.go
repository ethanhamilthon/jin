package headless

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHashPromptsAreExpandedLikeInTheTUI(t *testing.T) {
	var body string
	var h *harness
	h = newHarness(t, captureBody(&h, &body))
	h.dir = t.TempDir()
	prompts := filepath.Join(os.Getenv("HOME"), ".jin-dev", "prompts")
	if err := os.MkdirAll(prompts, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(prompts, "deploy.md"), []byte("Branch {{echo main}}"), 0o600); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		prompt string
		want   []string
		absent []string
	}{
		{"#deploy now", []string{"pasted-prompts", `name=\"deploy\"`, "Branch main", "now"}, nil},
		{"#plan something", []string{"pasted-prompts", `name=\"plan\"`}, nil},
		{"#nosuch something", []string{"#nosuch something"}, []string{"pasted-prompts"}},
		{"no references", []string{"no references"}, []string{"pasted-prompts"}},
	}
	for _, c := range cases {
		if code := h.run(t, "-p", "--no-session", c.prompt); code != 0 {
			t.Fatalf("%q: code %d, stderr %q", c.prompt, code, h.errOut.String())
		}
		for _, want := range c.want {
			if !strings.Contains(body, want) {
				t.Errorf("%q: body lacks %q", c.prompt, want)
			}
		}
		for _, no := range c.absent {
			if strings.Contains(body, no) {
				t.Errorf("%q: body has %q", c.prompt, no)
			}
		}
	}
}

func TestDisabledPromptIsNotExpandedInHeadless(t *testing.T) {
	var body string
	var h *harness
	h = newHarness(t, captureBody(&h, &body))
	h.dir = t.TempDir()
	if err := h.db.SavePromptsDisabled([]string{"plan"}); err != nil {
		t.Fatal(err)
	}
	h.run(t, "-p", "--no-session", "#plan x")
	if strings.Contains(body, "pasted-prompts") {
		t.Errorf("disabled prompt expanded")
	}
}
