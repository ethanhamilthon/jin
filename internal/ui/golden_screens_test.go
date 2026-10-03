package ui

import (
	"testing"

	"github.com/gdamore/tcell/v3"
	"github.com/gdamore/tcell/v3/vt"

	"jin/internal/core"
	"jin/internal/store"
	"jin/internal/tools"
)

// goldenApp is a fixed 72×28 screen with a session named "golden".
func goldenApp(t *testing.T) (*app, tcell.Screen) {
	t.Helper()
	t.Setenv("HOME", t.TempDir())
	screen, err := tcell.NewTerminfoScreenFromTty(vt.NewMockTerm(vt.MockOptSize{X: 72, Y: 28}))
	if err != nil {
		t.Fatal(err)
	}
	if err := screen.Init(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(screen.Fini)
	t.Cleanup(func() { applyTheme(themes[0]) })
	a := &app{screen: screen, width: 72, sessions: map[string]*chatSession{}, registry: tools.NewRegistry(), dir: "/work/jin", version: "v0.6", cfg: readyConfig()}
	a.active = &chatSession{id: "g", width: 72, model: "gpt-x", effort: "high", title: "golden", ready: true}
	a.sessions["g"] = a.active
	return a, screen
}

// conversation is a short exchange with every kind of row the chat draws.
func conversation(s *chatSession) {
	s.appendEntry(chatEntry{kind: core.UpdateUser, text: "Fix the failing test"})
	s.appendEntry(chatEntry{kind: core.UpdateReasoning, text: "The test expects a trimmed name."})
	s.appendEntry(chatEntry{kind: core.UpdateToolCall, tool: "bash", text: "go test ./... (120s)"})
	s.appendEntry(chatEntry{kind: core.UpdateToolResult, tool: "bash", text: resultText("bash", "--- FAIL: TestName\nname = \" a \"\nFAIL", nil)})
	s.appendEntry(chatEntry{kind: core.UpdateToolCall, tool: "edit", text: "name.go:12"})
	s.appendEntry(chatEntry{kind: core.UpdateToolResult, tool: "edit", text: resultText("edit", "", []tools.Change{{Before: "return n", After: "return strings.TrimSpace(n)"}})})
	s.appendEntry(chatEntry{kind: core.UpdateAssistant, text: "## Fixed\n\nThe name is **trimmed** now, see `name.go`.\n8. tests pass\n9. [docs](https://go.dev)\n\n| file | change |\n|---|---|\n| name.go | trim |"})
}

func TestGoldenIntro(t *testing.T) {
	a, screen := goldenApp(t)
	for _, entry := range a.introEntries() {
		a.active.appendEntry(entry)
	}
	a.draw()
	golden(t, screen, "intro")
}

func TestGoldenOnboarding(t *testing.T) {
	a, screen := goldenApp(t)
	a.cfg = store.Config{}
	a.draw()
	golden(t, screen, "onboarding")
	a.addProviderOfKind(onboardingKinds[0].value)
	a.draw()
	golden(t, screen, "onboarding-setup")
}

func TestGoldenChatInEveryFoldMode(t *testing.T) {
	for _, mode := range []struct {
		name string
		fold foldMode
	}{{"output", foldOutput}, {"all", foldAll}, {"no-tools", foldNoTools}, {"messages", foldMessages}} {
		a, screen := goldenApp(t)
		conversation(a.active)
		a.active.setFold(mode.fold)
		a.draw()
		golden(t, screen, "chat-"+mode.name)
	}
}

func TestGoldenThemes(t *testing.T) {
	for _, th := range themes {
		a, screen := goldenApp(t)
		applyTheme(th)
		conversation(a.active)
		a.draw()
		golden(t, screen, "theme-"+goldenName(th.name))
	}
}

func goldenName(name string) string {
	out := []rune{}
	for _, r := range name {
		switch {
		case r >= 'A' && r <= 'Z':
			out = append(out, r+'a'-'A')
		case r >= 'a' && r <= 'z':
			out = append(out, r)
		case r == ' ':
			out = append(out, '-')
		case r == 'é':
			out = append(out, 'e')
		}
	}
	return string(out)
}
