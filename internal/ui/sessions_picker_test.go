package ui

import (
	"slices"
	"testing"

	"github.com/gdamore/tcell/v3"

	"jin/internal/tools"
)

// shownLabels lists the labels a panel draws, group headers included.
func shownLabels(sel *selector) []string {
	var labels []string
	for _, i := range sel.rows() {
		labels = append(labels, sel.options[i].label)
	}
	return labels
}

func TestSessionStepGroupsRunningSessionsFirst(t *testing.T) {
	a := startingApp(t)
	dir := registerProject(t, a, "app")
	if err := a.store.Touch("run", dir, "m", "high", "Refactor parser"); err != nil {
		t.Fatal(err)
	}
	if err := a.store.SetRunning("run", true); err != nil {
		t.Fatal(err)
	}
	if err := a.store.Touch("idle", dir, "m", "high", "Fix login"); err != nil {
		t.Fatal(err)
	}
	if err := a.store.Touch("docs", dir, "m", "high", "Write docs"); err != nil {
		t.Fatal(err)
	}
	a.openProjectSessions(dir)
	labels := shownLabels(a.sel)
	if len(labels) != 6 || !slices.Equal(labels[:4], []string{"New session", "Running", "Refactor parser", "Last hour"}) {
		t.Fatalf("rows = %q, want New session, Running, its session, Last hour, then the other two", labels)
	}
	if !slices.Contains(labels[4:], "Fix login") || !slices.Contains(labels[4:], "Write docs") {
		t.Fatalf("rows = %q, want the idle sessions under Last hour", labels)
	}
}

func TestSessionStepSearchMatchesTitlesAndHidesEmptyHeaders(t *testing.T) {
	a := startingApp(t)
	dir := registerProject(t, a, "app")
	if err := a.store.Touch("run", dir, "m", "high", "Refactor parser"); err != nil {
		t.Fatal(err)
	}
	if err := a.store.SetRunning("run", true); err != nil {
		t.Fatal(err)
	}
	if err := a.store.Touch("idle", dir, "m", "high", "Fix login"); err != nil {
		t.Fatal(err)
	}
	a.openProjectSessions(dir)
	for _, r := range []string{"F", "I", "X"} {
		typeRune(a, r)
	}
	if got := shownLabels(a.sel); !slices.Equal(got, []string{"Last hour", "Fix login"}) {
		t.Fatalf("search rows = %q, want the Last hour header and Fix login", got)
	}
}

func TestSessionStepEscAndLeftGoBackToProjects(t *testing.T) {
	a := startingApp(t)
	dir := registerProject(t, a, "app")
	a.openProjects()
	a.sel.selectValue(dir)
	a.submitSelector()
	a.selectorKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	if a.sel == nil || a.sel.title != "Projects" || a.sel.current() != dir {
		t.Fatalf("Esc = %+v, want the projects with the project under the cursor", a.sel)
	}
	a.submitSelector()
	a.selectorKey(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone))
	if a.sel == nil || a.sel.title != "Projects" {
		t.Fatalf("Left with an empty search = %+v, want the projects", a.sel)
	}
	a.submitSelector()
	typeRune(a, "x")
	a.selectorKey(tcell.NewEventKey(tcell.KeyLeft, "", tcell.ModNone))
	if a.sel == nil || a.sel.title == "Projects" {
		t.Fatal("Left in a search must move the cursor, not leave the step")
	}
}

func TestNewSessionRowStartsInTheChosenProject(t *testing.T) {
	a := startingApp(t)
	dir := registerProject(t, a, "app")
	a.openProjectSessions(dir)
	a.submitSelector()
	if a.active == nil || a.active.path != dir {
		t.Fatalf("active path = %v, want %q", a.active, dir)
	}
}

func TestSessionRowResumesTheSession(t *testing.T) {
	a := startingApp(t)
	a.registry = tools.NewRegistry()
	dir := registerProject(t, a, "app")
	if err := a.store.Touch("idle", dir, "m", "high", "Fix login"); err != nil {
		t.Fatal(err)
	}
	a.openProjectSessions(dir)
	a.sel.selectValue("idle")
	a.submitSelector()
	if a.active == nil || a.active.id != "idle" || a.active.path != dir {
		t.Fatalf("active session = %+v, want idle in %q", a.active, dir)
	}
}

func TestSessionsCommandOpensTheCurrentProjectStep(t *testing.T) {
	a := startingApp(t)
	cmd, ok := slashByName("sessions")
	if !ok {
		t.Fatal("/sessions is missing")
	}
	cmd.run(a, "")
	if a.sel == nil || a.sel.title != "Sessions · "+shortPath(a.dir) {
		t.Fatalf("/sessions = %+v, want the sessions of the current project", a.sel)
	}
	a.selectorKey(tcell.NewEventKey(tcell.KeyEscape, "", tcell.ModNone))
	if a.sel == nil || a.sel.title != "Projects" {
		t.Fatalf("Esc in /sessions = %+v, want the projects", a.sel)
	}
}
