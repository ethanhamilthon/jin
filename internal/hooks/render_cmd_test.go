package hooks

import (
	"bytes"
	"context"
	"os"
	"strings"
	"testing"
)

func runRender(t *testing.T, dir string, settings Settings) (string, string) {
	t.Helper()
	var out, errOut bytes.Buffer
	if code := RenderFilled(context.Background(), dir, settings, &out, &errOut); code != 0 {
		t.Fatalf("exit %d", code)
	}
	return out.String(), errOut.String()
}

func TestRenderFilledFillsAndSkips(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	for name, body := range map[string]string{"10-a": "Branch: {{echo main}}", "20-off": "Hidden.", "30-b": "Broken: {{exit 3}}"} {
		if _, err := Create(name); err != nil {
			t.Fatal(err)
		}
		write(t, name, body)
	}
	out, errOut := runRender(t, t.TempDir(), Settings{Disabled: []string{"20-off"}})
	if want := "Branch: main\n\nBroken: [command failed: exit status 3]"; strings.TrimSpace(out) != want {
		t.Errorf("out = %q, want %q", out, want)
	}
	if !strings.Contains(errOut, "exit status 3") || strings.Contains(out, "Hidden") {
		t.Errorf("errOut = %q, out = %q", errOut, out)
	}
}

func TestRenderFilledProjectHooksNeedTrust(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	path, err := CreateProject(dir, "lint")
	if err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(path, []byte("Run make lint."), 0o644)
	out, errOut := runRender(t, dir, Settings{})
	if out != "" || !strings.Contains(errOut, "not trusted") {
		t.Errorf("untrusted: out = %q, errOut = %q", out, errOut)
	}
	if out, _ := runRender(t, dir, Settings{Trusted: true}); strings.TrimSpace(out) != "Run make lint." {
		t.Errorf("trusted: out = %q", out)
	}
}
