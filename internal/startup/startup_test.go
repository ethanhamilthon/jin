package startup

import (
	"context"
	"jin/internal/provider"
	"jin/internal/testjin"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func setup(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	return filepath.Join(home, ".jin-dev")
}

func write(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o600); err != nil {
		t.Fatal(err)
	}
}

func render(t *testing.T, in Input) Output {
	t.Helper()
	if in.Dir == "" {
		in.Dir = t.TempDir()
	}
	if in.Env == nil {
		in.Env = fakeJin(t)
	}
	return Render(context.Background(), in, nil)
}

func fakeJin(t *testing.T) []string { return testjin.Env(t) }

func TestBuiltInSystemPromptExpandsItsCommands(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	write(t, filepath.Join(dir, "AGENTS.md"), "Project rules.")
	out := render(t, Input{Dir: dir, SessionID: "sess-9", Env: fakeJin(t)})
	if strings.Contains(out.System, "{{") {
		t.Fatalf("a placeholder was left in the prompt:\n%s", out.System)
	}
	for _, want := range []string{
		"Jin documentation: pointer", "Hook text", "AGENTS.md:\n\nProject rules.",
		"Working directory: ", "- OS: ", "Date: 20", "Your session id: sess-9",
	} {
		if !strings.Contains(out.System, want) {
			t.Errorf("prompt lacks %q:\n%s", want, out.System)
		}
	}
	if strings.Index(out.System, "Project rules.") > strings.Index(out.System, provider.CacheBreak) || strings.Index(out.System, "Working directory") < strings.Index(out.System, provider.CacheBreak) {
		t.Errorf("stable text must come before the cache break, live data after it:\n%s", out.System)
	}
	if out.Compact == "" || out.Handoff == "" || len(out.Warnings) != 0 {
		t.Errorf("output = %+v", out)
	}
}

func TestCustomSystemPromptCommandsNeverRun(t *testing.T) {
	root := setup(t)
	sentinel := filepath.Join(t.TempDir(), "custom-ran")
	write(t, filepath.Join(root, "system-prompt.md"), "# system\nCUSTOM_SENTINEL {{touch "+sentinel+"}}\n# compact\nCUSTOM_COMPACT\n# handoff\nCUSTOM_HANDOFF\n")
	out := render(t, Input{})
	if strings.Contains(out.System, "CUSTOM_SENTINEL") || strings.Contains(out.Compact, "CUSTOM_COMPACT") || strings.Contains(out.Handoff, "CUSTOM_HANDOFF") {
		t.Fatalf("custom prompt reached output: %+v", out)
	}
	if _, err := os.Stat(sentinel); !os.IsNotExist(err) {
		t.Fatalf("custom prompt command ran, stat error = %v", err)
	}
	path := filepath.Join(root, "system-prompt.md")
	if data, err := os.ReadFile(path); err != nil || !strings.Contains(string(data), "CUSTOM_SENTINEL") {
		t.Fatalf("custom file was not preserved: %q, %v", data, err)
	}
}

func TestReusablePromptsCommandsStillRun(t *testing.T) {
	root := setup(t)
	write(t, filepath.Join(root, "prompts", "deploy.md"), "Branch: {{echo main}}")
	out := render(t, Input{WithPrompts: true})
	if out.Prompts["deploy"] != "Branch: main" {
		t.Errorf("deploy = %q", out.Prompts["deploy"])
	}
	if out.Prompts["plan"] == "" {
		t.Error("the built-in prompts must be there")
	}
}

func TestPromptsAreSkippedWhenNotAskedForAndWhenDisabled(t *testing.T) {
	root := setup(t)
	write(t, filepath.Join(root, "prompts", "off.md"), "off text")
	if out := render(t, Input{}); len(out.Prompts) != 0 {
		t.Errorf("prompts = %v, want none without WithPrompts", out.Prompts)
	}
	out := render(t, Input{WithPrompts: true, PromptsDisabled: []string{"off", "plan"}})
	if _, ok := out.Prompts["off"]; ok {
		t.Error("a disabled prompt must not be rendered")
	}
	if _, ok := out.Prompts["plan"]; ok {
		t.Error("a disabled system prompt must not be rendered")
	}
}

// The point of the trust boundary: AGENTS.md comes from a repository, and a
// repository must not be able to run code just because jin was opened in it.
func TestAgentsMdIsNeverRun(t *testing.T) {
	setup(t)
	dir := t.TempDir()
	marker := filepath.Join(t.TempDir(), "ran")
	write(t, filepath.Join(dir, "AGENTS.md"), "Rules {{touch "+marker+"}} and {{echo SURPRISE}}")
	out := render(t, Input{Dir: dir, Env: fakeJin(t)})
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("a command in AGENTS.md was run")
	}
	if strings.Contains(out.System, "SURPRISE") && !strings.Contains(out.System, "{{echo SURPRISE}}") {
		t.Errorf("AGENTS.md was expanded:\n%s", out.System)
	}
	if !strings.Contains(out.System, "{{touch "+marker+"}} and {{echo SURPRISE}}") {
		t.Errorf("AGENTS.md text was changed:\n%s", out.System)
	}
}

func TestFailingHookCommandsGiveWarningsAndTheRestStillWorks(t *testing.T) {
	root := setup(t)
	write(t, filepath.Join(root, "prompts", "bad.md"), "Good {{echo fine}} bad {{exit 4}}")
	out := render(t, Input{WithPrompts: true})
	if out.Prompts["bad"] != "Good fine bad [command failed: exit status 4]" {
		t.Errorf("prompt = %q", out.Prompts["bad"])
	}
	if len(out.Warnings) != 1 || !strings.Contains(out.Warnings[0], "exit status 4") {
		t.Errorf("warnings = %v", out.Warnings)
	}
}

func TestOnPromptIsCalledOncePerPromptAsEachOneIsDone(t *testing.T) {
	root := setup(t)
	write(t, filepath.Join(root, "prompts", "fast.md"), "fast {{echo f}}")
	write(t, filepath.Join(root, "prompts", "slow.md"), "slow {{sleep 0.6; echo s}}")
	var mu sync.Mutex
	var order []string
	var times = map[string]time.Time{}
	start := time.Now()
	out := Render(context.Background(), Input{Dir: t.TempDir(), Env: fakeJin(t), WithPrompts: true, PromptsDisabled: []string{"plan", "review"}}, func(name string) {
		mu.Lock()
		order = append(order, name)
		times[name] = time.Now()
		mu.Unlock()
	})
	if len(order) != 2 || order[0] != "fast" || order[1] != "slow" {
		t.Fatalf("onPrompt order = %v, want fast then slow", order)
	}
	if times["fast"].Sub(start) > 400*time.Millisecond {
		t.Errorf("the fast prompt was reported after %v: it waited for the slow one", times["fast"].Sub(start))
	}
	if out.Prompts["slow"] != "slow s" || out.Prompts["fast"] != "fast f" {
		t.Errorf("prompts = %v", out.Prompts)
	}
}

func TestHookAndPromptCommandsRunAtTheSameTime(t *testing.T) {
	root := setup(t)
	write(t, filepath.Join(root, "hooks", "h.md"), "{{sleep 0.5}}b")
	write(t, filepath.Join(root, "prompts", "p.md"), "{{sleep 0.5}}c")
	start := time.Now()
	render(t, Input{WithPrompts: true, PromptsDisabled: []string{"plan", "review"}})
	if took := time.Since(start); took > 1400*time.Millisecond {
		t.Errorf("two half-second commands took %v: they did not run together", took)
	}
}

func TestCancelStopsTheCommandsAndStillReturnsAPrompt(t *testing.T) {
	root := setup(t)
	write(t, filepath.Join(root, "prompts", "slow.md"), "waiting {{sleep 30}}")
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan Output, 1)
	go func() {
		done <- Render(ctx, Input{Dir: t.TempDir(), Env: fakeJin(t), WithPrompts: true, PromptsDisabled: []string{"plan", "review"}}, nil)
	}()
	time.Sleep(300 * time.Millisecond)
	cancel()
	select {
	case out := <-done:
		if out.Prompts["slow"] != "waiting [command cancelled]" {
			t.Errorf("slow = %q", out.Prompts["slow"])
		}
		if !strings.Contains(out.System, "You are jin") {
			t.Errorf("the static prompt must survive cancelled commands:\n%s", out.System)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("cancel did not end the render")
	}
}

func TestCommandsSeeTheSessionAndThisJin(t *testing.T) {
	hooks := setup(t)
	write(t, filepath.Join(hooks, "prompts", "context.md"), "id={{echo $JIN_SESSION_ID}} dir={{echo $JIN_DIR}} path={{echo $PATH}}\n")
	dir := t.TempDir()
	out := render(t, Input{Dir: dir, SessionID: "abc123", WithPrompts: true})
	exe, _ := os.Executable()
	for _, want := range []string{"id=abc123", "dir=" + dir, "path=" + filepath.Dir(exe) + string(os.PathListSeparator)} {
		if !strings.Contains(out.Prompts["context"], want) {
			t.Errorf("system prompt lacks %q:\n%s", want, out.Prompts["context"])
		}
	}
}
