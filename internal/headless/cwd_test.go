package headless

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCwdFlagSetsSessionPathAndProjectFiles(t *testing.T) {
	var body string
	var h *harness
	h = newHarness(t, captureBody(&h, &body))
	t.Chdir(t.TempDir())
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "AGENTS.md"), []byte("PROJECT-RULE-42"), 0o600); err != nil {
		t.Fatal(err)
	}
	h.dir = t.TempDir()
	if code := h.run(t, "-p", "--cwd", project, "hi"); code != 0 {
		t.Fatalf("code %d %s", code, h.errOut.String())
	}
	want, _ := filepath.EvalSymlinks(project)
	list, _ := h.db.ListByPath(project)
	if len(list) != 1 {
		t.Fatalf("no session stored for %s", project)
	}
	if got, _ := os.Getwd(); got != want {
		t.Errorf("process cwd %q, want %q", got, want)
	}
	if !strings.Contains(body, "PROJECT-RULE-42") {
		t.Error("AGENTS.md of --cwd was not read")
	}
	if code := h.run(t, "-p", "--cwd", filepath.Join(project, "AGENTS.md"), "hi"); code != exitError || !strings.Contains(h.errOut.String(), "is not a directory") {
		t.Errorf("file as cwd: code %d %q", code, h.errOut.String())
	}
}
