package agentkit

import (
	"context"
	"slices"
	"testing"

	"jin/internal/core"
	"jin/internal/provider"
	"jin/internal/startup"
	"jin/internal/tools"
)

func build(t *testing.T, mode Mode, names []string, dir string) (*core.Agent, startup.Input) {
	t.Helper()
	in := startup.Input{Dir: dir, SessionID: "s1"}
	agent := New(Spec{
		Mode: mode, Client: provider.NewClient(provider.Config{}), Names: names,
		Dir: dir, Owner: "s1",
	})
	Apply(agent, startup.Render(context.Background(), in, nil))
	return agent, in
}

func TestModesHaveTheSameToolsExceptBashTimeoutText(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	withoutBash := slices.DeleteFunc(tools.Catalog(), func(name string) bool { return name == "bash" })
	interactive, _ := build(t, Interactive, withoutBash, dir)
	oneShot, _ := build(t, OneShot, withoutBash, dir)
	if interactive.ToolSchemaBytes() != oneShot.ToolSchemaBytes() {
		t.Fatalf("tool schemas differ without bash: %d and %d", interactive.ToolSchemaBytes(), oneShot.ToolSchemaBytes())
	}
	interactive, _ = build(t, Interactive, tools.Catalog(), dir)
	oneShot, _ = build(t, OneShot, tools.Catalog(), dir)
	if interactive.ToolSchemaBytes() == oneShot.ToolSchemaBytes() {
		t.Fatal("the headless bash tool should say that a command past its timeout is killed")
	}
}

func TestModesBuildTheSameSystemPrompt(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	dir := t.TempDir()
	interactive, _ := build(t, Interactive, tools.Catalog(), dir)
	oneShot, _ := build(t, OneShot, tools.Catalog(), dir)
	if interactive.SystemPrompt() == "" {
		t.Fatal("Apply left the system prompt empty")
	}
	if interactive.SystemPrompt() != oneShot.SystemPrompt() {
		t.Fatal("interactive and one-shot agents got different system prompts")
	}
}

func TestRefresherRendersTheSystemPromptOfApply(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	agent, in := build(t, Interactive, tools.Catalog(), t.TempDir())
	if got := Refresher(in)(context.Background()); got != agent.SystemPrompt() {
		t.Fatal("the refresher renders another system prompt than the one that Apply set")
	}
}
